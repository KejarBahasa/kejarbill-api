package integration_test

import (
	"context"
	"errors"
	"testing"

	expenseConstants "github.com/KejarBahasa/kejarbill-api/internal/module/expense/constants"
	expenseDto "github.com/KejarBahasa/kejarbill-api/internal/module/expense/dto"
	expenseRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/expense/repository"
	expenseServicePkg "github.com/KejarBahasa/kejarbill-api/internal/module/expense/service"
	groupRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/group/repository"
	groupMemberRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/group_member/repository"
	groupParticipantRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/group_participant/repository"
	ledgerRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/ledger/repository"
	ledgerServicePkg "github.com/KejarBahasa/kejarbill-api/internal/module/ledger/service"
	settlementRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/settlement/repository"
	"github.com/google/uuid"
)

func newItemizedService() *expenseServicePkg.ExpenseService {
	return expenseServicePkg.NewExpenseService(
		integrationPool,
		githubExpenseRepository(),
		ledgerRepoPkg.NewLedgerRepository(),
		ledgerServicePkg.NewLedgerService(),
		groupRepoPkg.NewGroupRepository(),
		groupMemberRepoPkg.NewGroupMemberRepository(),
		groupParticipantRepoPkg.NewGroupParticipantRepository(),
		settlementRepoPkg.NewSettlementRepository(),
	)
}

func githubExpenseRepository() *expenseRepoPkg.ExpenseRepository {
	return expenseRepoPkg.NewExpenseRepository()
}

func itemizedRequest(f *equalExpenseFixture, items []expenseDto.CreateExpenseItemizedItem) *expenseDto.CreateExpenseItemizedRequest {
	return &expenseDto.CreateExpenseItemizedRequest{
		GroupID:            f.groupID,
		Title:              "Indomaret",
		Currency:           "IDR",
		ExpenseDate:        "2026-08-13T00:00:00Z",
		PayerParticipantID: f.payerParticipant,
		Items:              items,
	}
}

func TestItemizedExpenseSupportsIndividualAndSharedItems(t *testing.T) {
	f := newEqualExpenseFixture(t)
	service := newItemizedService()

	expenseID, err := service.CreateItemizedExpense(context.Background(), f.ownerID, itemizedRequest(f, []expenseDto.CreateExpenseItemizedItem{
		{Name: "Roti", Qty: 1, UnitPrice: 10_000, ParticipantIDs: []string{f.guestParticipant}},
		{Name: "Coca-Cola 1L", Qty: 1, UnitPrice: 20_000, ParticipantIDs: []string{f.payerParticipant, f.guestParticipant, f.thirdParticipant}},
	}))
	if err != nil {
		t.Fatalf("CreateItemizedExpense() error = %v", err)
	}

	var totalShare int64
	if err := integrationPool.QueryRow(context.Background(), `
		SELECT COALESCE(SUM(share_amount), 0)
		FROM expense_item_participants eip
		JOIN expense_items ei ON ei.id = eip.expense_item_id
		WHERE ei.expense_id = $1
	`, expenseID).Scan(&totalShare); err != nil {
		t.Fatalf("query item shares: %v", err)
	}
	if totalShare != 30_000 {
		t.Fatalf("total item share = %d, want 30000", totalShare)
	}

	detail, err := service.GetDetailByID(context.Background(), f.ownerID, expenseID)
	if err != nil {
		t.Fatalf("GetDetailByID() error = %v", err)
	}
	if len(detail.Items) != 2 || len(detail.Items[0].Participants) != 1 || len(detail.Items[1].Participants) != 3 {
		t.Fatalf("item detail participants = %+v, want 1 and 3", detail.Items)
	}
	shared := detail.Items[1].Participants
	sharesByParticipant := make(map[string]int64, len(shared))
	for _, participant := range shared {
		sharesByParticipant[participant.ParticipantID] = participant.ShareAmount
	}
	if sharesByParticipant[f.payerParticipant] != 6_667 || sharesByParticipant[f.guestParticipant] != 6_667 || sharesByParticipant[f.thirdParticipant] != 6_666 {
		t.Fatalf("shared item shares = %+v, want payer=6667 guest=6667 third=6666", shared)
	}

	var guestLiability, thirdLiability int64
	if err := integrationPool.QueryRow(context.Background(), `
		SELECT COALESCE(SUM(amount) FILTER (WHERE from_participant_id = $2), 0),
		       COALESCE(SUM(amount) FILTER (WHERE from_participant_id = $3), 0)
		FROM account_ledger
		WHERE group_id = $1 AND to_participant_id = $4 AND source_type = 'expense' AND source_id = $5
	`, f.groupID, f.guestParticipant, f.thirdParticipant, f.payerParticipant, expenseID).Scan(&guestLiability, &thirdLiability); err != nil {
		t.Fatalf("query ledger: %v", err)
	}
	if guestLiability != 16_667 || thirdLiability != 6_666 {
		t.Fatalf("ledger liabilities = %d, %d; want 16667, 6666", guestLiability, thirdLiability)
	}
}

func TestItemizedExpenseRejectsDuplicateParticipantPerItem(t *testing.T) {
	f := newEqualExpenseFixture(t)
	_, err := newItemizedService().CreateItemizedExpense(context.Background(), f.ownerID, itemizedRequest(f, []expenseDto.CreateExpenseItemizedItem{
		{Name: "Sauce", Qty: 1, UnitPrice: 10_000, ParticipantIDs: []string{f.guestParticipant, f.guestParticipant}},
	}))
	if !errors.Is(err, expenseConstants.ErrDuplicateParticipants) {
		t.Fatalf("error = %v, want %v", err, expenseConstants.ErrDuplicateParticipants)
	}
}

func TestItemizedExpenseRejectsParticipantFromAnotherGroup(t *testing.T) {
	f := newEqualExpenseFixture(t)
	foreignGroupID := uuid.NewString()
	foreignParticipantID := uuid.NewString()
	mustExec(t, context.Background(), `INSERT INTO groups (id, name, created_by) VALUES ($1, 'Foreign Item Group', $2)`, foreignGroupID, f.ownerID)
	mustExec(t, context.Background(), `
		INSERT INTO group_participants (id, group_id, display_name, participant_type, created_by)
		VALUES ($1, $2, 'Foreign', 'guest', $3)
	`, foreignParticipantID, foreignGroupID, f.ownerID)
	t.Cleanup(func() {
		mustExec(t, context.Background(), `DELETE FROM group_participants WHERE id = $1`, foreignParticipantID)
		mustExec(t, context.Background(), `DELETE FROM groups WHERE id = $1`, foreignGroupID)
	})

	_, err := newItemizedService().CreateItemizedExpense(context.Background(), f.ownerID, itemizedRequest(f, []expenseDto.CreateExpenseItemizedItem{
		{Name: "Foreign", Qty: 1, UnitPrice: 10_000, ParticipantIDs: []string{f.payerParticipant, foreignParticipantID}},
	}))
	if !errors.Is(err, expenseConstants.ErrParticipantNotInGroup) {
		t.Fatalf("error = %v, want %v", err, expenseConstants.ErrParticipantNotInGroup)
	}
}

func TestLegacyItemizedExpenseWithoutItemRelationsRemainsReadable(t *testing.T) {
	f := newEqualExpenseFixture(t)
	service := newItemizedService()
	expenseID := uuid.NewString()
	itemID := uuid.NewString()
	mustExec(t, context.Background(), `
		INSERT INTO expenses (id, group_id, title, paid_by_participant_id, total_amount, split_method, expense_date, created_by)
		VALUES ($1, $2, 'Legacy itemized', $3, 10000, 'itemized', '2026-08-13T00:00:00Z', $4)
	`, expenseID, f.groupID, f.payerParticipant, f.ownerID)
	mustExec(t, context.Background(), `
		INSERT INTO expense_items (id, expense_id, name, qty, unit_price, subtotal)
		VALUES ($1, $2, 'Legacy bread', 1, 10000, 10000)
	`, itemID, expenseID)

	detail, err := service.GetDetailByID(context.Background(), f.ownerID, expenseID)
	if err != nil {
		t.Fatalf("GetDetailByID() error = %v", err)
	}
	if len(detail.Items) != 1 || detail.Items[0].ID != itemID || detail.Items[0].Name != "Legacy bread" {
		t.Fatalf("legacy items = %+v, want legacy item readable", detail.Items)
	}
}
