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

type equalExpenseFixture struct {
	service *expenseServicePkg.ExpenseService

	groupID          string
	ownerID          string
	payerParticipant string
	guestParticipant string
	thirdParticipant string
}

func newEqualExpenseFixture(t *testing.T) *equalExpenseFixture {
	t.Helper()

	f := &equalExpenseFixture{
		service: expenseServicePkg.NewExpenseService(
			integrationPool,
			expenseRepoPkg.NewExpenseRepository(),
			ledgerRepoPkg.NewLedgerRepository(),
			ledgerServicePkg.NewLedgerService(),
			groupRepoPkg.NewGroupRepository(),
			groupMemberRepoPkg.NewGroupMemberRepository(),
			groupParticipantRepoPkg.NewGroupParticipantRepository(),
			settlementRepoPkg.NewSettlementRepository(),
		),
		groupID:          uuid.NewString(),
		ownerID:          uuid.NewString(),
		payerParticipant: uuid.NewString(),
		guestParticipant: uuid.NewString(),
		thirdParticipant: uuid.NewString(),
	}

	ctx := context.Background()
	mustExec(t, ctx, `
		INSERT INTO users (id, email, username, full_name, password_hash)
		VALUES ($1, $2, $3, 'Equal Test Owner', 'not-a-real-password')
	`, f.ownerID, f.ownerID+"@example.test", "eq_"+f.ownerID[:8])
	mustExec(t, ctx, `
		INSERT INTO groups (id, name, created_by) VALUES ($1, 'Equal Test Group', $2)
	`, f.groupID, f.ownerID)
	mustExec(t, ctx, `
		INSERT INTO group_members (group_id, user_id, role, status) VALUES ($1, $2, 'owner', 'active')
	`, f.groupID, f.ownerID)
	mustExec(t, ctx, `
		INSERT INTO group_participants (id, group_id, user_id, display_name, participant_type, created_by)
		VALUES ($1, $2, $3, 'Payer', 'registered', $3)
	`, f.payerParticipant, f.groupID, f.ownerID)
	mustExec(t, ctx, `
		INSERT INTO group_participants (id, group_id, user_id, display_name, participant_type, created_by)
		VALUES ($1, $2, NULL, 'Guest', 'guest', $3)
	`, f.guestParticipant, f.groupID, f.ownerID)
	mustExec(t, ctx, `
		INSERT INTO group_participants (id, group_id, user_id, display_name, participant_type, created_by)
		VALUES ($1, $2, NULL, 'Guest Two', 'guest', $3)
	`, f.thirdParticipant, f.groupID, f.ownerID)

	t.Cleanup(func() { f.cleanup(t) })
	return f
}

func (f *equalExpenseFixture) cleanup(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	statements := []struct {
		query string
		args  []any
	}{
		{`DELETE FROM account_ledger WHERE group_id = $1`, []any{f.groupID}},
		{`DELETE FROM expense_participants WHERE expense_id IN (SELECT id FROM expenses WHERE group_id = $1)`, []any{f.groupID}},
		{`DELETE FROM expense_item_participants WHERE expense_item_id IN (SELECT ei.id FROM expense_items ei JOIN expenses e ON e.id = ei.expense_id WHERE e.group_id = $1)`, []any{f.groupID}},
		{`DELETE FROM expense_item_participants WHERE expense_item_id IN (SELECT ei.id FROM expense_items ei JOIN expenses e ON e.id = ei.expense_id WHERE e.group_id = $1)`, []any{f.groupID}},
		{`DELETE FROM expense_items WHERE expense_id IN (SELECT id FROM expenses WHERE group_id = $1)`, []any{f.groupID}},
		{`DELETE FROM settlements WHERE group_id = $1`, []any{f.groupID}},
		{`DELETE FROM expenses WHERE group_id = $1`, []any{f.groupID}},
		{`DELETE FROM group_participants WHERE group_id = $1`, []any{f.groupID}},
		{`DELETE FROM group_members WHERE group_id = $1`, []any{f.groupID}},
		{`DELETE FROM groups WHERE id = $1`, []any{f.groupID}},
		{`DELETE FROM users WHERE id = ANY($1)`, []any{[]string{f.ownerID}}},
	}
	for _, s := range statements {
		if _, err := integrationPool.Exec(ctx, s.query, s.args...); err != nil {
			t.Errorf("cleanup failed: %v", err)
		}
	}
}

func (f *equalExpenseFixture) equalReq(totalAmount int64) *expenseDto.CreateExpenseEqualRequest {
	return f.equalReqWithParticipants(totalAmount, []string{f.payerParticipant, f.guestParticipant})
}

func (f *equalExpenseFixture) equalReqWithParticipants(totalAmount int64, participantIDs []string) *expenseDto.CreateExpenseEqualRequest {
	return &expenseDto.CreateExpenseEqualRequest{
		GroupID:            f.groupID,
		Title:              "Equal test",
		Currency:           "IDR",
		PayerParticipantID: f.payerParticipant,
		ParticipantIDs:     participantIDs,
		TotalAmount:        totalAmount,
		ExpenseDate:        "2026-08-13T00:00:00Z",
	}
}

func (f *equalExpenseFixture) expenseCount(t *testing.T) int {
	t.Helper()
	return countRows(t, context.Background(), `SELECT COUNT(*) FROM expenses WHERE group_id = $1`, f.groupID)
}

func TestEqualExpenseDistributesRemainderInRequestOrder(t *testing.T) {
	fixture := newEqualExpenseFixture(t)

	expenseID, err := fixture.service.CreateEqualExpense(context.Background(), fixture.ownerID, fixture.equalReq(101))
	if err != nil {
		t.Fatalf("CreateEqualExpense() error = %v", err)
	}

	var firstShare, secondShare, totalShare int64
	if err := integrationPool.QueryRow(context.Background(), `
		SELECT
			SUM(CASE WHEN ep.participant_id = $2 THEN ep.share_amount ELSE 0 END),
			SUM(CASE WHEN ep.participant_id = $3 THEN ep.share_amount ELSE 0 END),
			SUM(ep.share_amount)
		FROM expense_participants ep
		JOIN expenses e ON e.id = ep.expense_id
		WHERE e.group_id = $1
	`, fixture.groupID, fixture.payerParticipant, fixture.guestParticipant).Scan(&firstShare, &secondShare, &totalShare); err != nil {
		t.Fatalf("query shares: %v", err)
	}
	if firstShare != 51 || secondShare != 50 || totalShare != 101 {
		t.Fatalf("shares = %d, %d, total %d; want 51, 50, 101", firstShare, secondShare, totalShare)
	}

	detail, err := fixture.service.GetDetailByID(context.Background(), fixture.ownerID, expenseID)
	if err != nil {
		t.Fatalf("GetDetailByID() error = %v", err)
	}
	sharesByParticipant := make(map[string]int64, len(detail.Participants))
	for _, participant := range detail.Participants {
		sharesByParticipant[participant.ParticipantID] = participant.ShareAmount
	}
	if len(detail.Participants) != 2 || sharesByParticipant[fixture.payerParticipant] != 51 || sharesByParticipant[fixture.guestParticipant] != 50 {
		t.Fatalf("detail participants = %+v, want shares 51 and 50", detail.Participants)
	}
}

func TestEqualExpenseAcceptsDivisible(t *testing.T) {
	fixture := newEqualExpenseFixture(t)

	if _, err := fixture.service.CreateEqualExpense(context.Background(), fixture.ownerID, fixture.equalReq(100)); err != nil {
		t.Fatalf("CreateEqualExpense() error = %v", err)
	}
	if n := fixture.expenseCount(t); n != 1 {
		t.Fatalf("expense count = %d, want 1", n)
	}

	var share int64
	if err := integrationPool.QueryRow(context.Background(), `
		SELECT share_amount FROM expense_participants ep
		JOIN expenses e ON e.id = ep.expense_id
		WHERE e.group_id = $1 AND ep.participant_id = $2
	`, fixture.groupID, fixture.guestParticipant).Scan(&share); err != nil {
		t.Fatalf("query share: %v", err)
	}
	if share != 50 {
		t.Fatalf("guest share = %d, want 50", share)
	}
}

func TestEqualExpenseDistributesRemainderAcrossThreeParticipants(t *testing.T) {
	fixture := newEqualExpenseFixture(t)
	req := fixture.equalReqWithParticipants(100_001, []string{fixture.payerParticipant, fixture.guestParticipant, fixture.thirdParticipant})
	if _, err := fixture.service.CreateEqualExpense(context.Background(), fixture.ownerID, req); err != nil {
		t.Fatalf("CreateEqualExpense() error = %v", err)
	}

	var firstShare, secondShare, thirdShare, totalShare int64
	if err := integrationPool.QueryRow(context.Background(), `
		SELECT
			SUM(CASE WHEN ep.participant_id = $2 THEN ep.share_amount ELSE 0 END),
			SUM(CASE WHEN ep.participant_id = $3 THEN ep.share_amount ELSE 0 END),
			SUM(CASE WHEN ep.participant_id = $4 THEN ep.share_amount ELSE 0 END),
			SUM(ep.share_amount)
		FROM expense_participants ep
		JOIN expenses e ON e.id = ep.expense_id
		WHERE e.group_id = $1
	`, fixture.groupID, fixture.payerParticipant, fixture.guestParticipant, fixture.thirdParticipant).Scan(&firstShare, &secondShare, &thirdShare, &totalShare); err != nil {
		t.Fatalf("query shares: %v", err)
	}
	if firstShare != 33_334 || secondShare != 33_334 || thirdShare != 33_333 || totalShare != 100_001 {
		t.Fatalf("shares = %d, %d, %d, total %d; want 33334, 33334, 33333, 100001", firstShare, secondShare, thirdShare, totalShare)
	}
}

func TestEqualExpenseDistributesSingleRemainderAcrossThreeParticipants(t *testing.T) {
	fixture := newEqualExpenseFixture(t)
	req := fixture.equalReqWithParticipants(100_000, []string{fixture.payerParticipant, fixture.guestParticipant, fixture.thirdParticipant})
	if _, err := fixture.service.CreateEqualExpense(context.Background(), fixture.ownerID, req); err != nil {
		t.Fatalf("CreateEqualExpense() error = %v", err)
	}

	var firstShare, secondShare, thirdShare int64
	if err := integrationPool.QueryRow(context.Background(), `
		SELECT
			SUM(CASE WHEN ep.participant_id = $2 THEN ep.share_amount ELSE 0 END),
			SUM(CASE WHEN ep.participant_id = $3 THEN ep.share_amount ELSE 0 END),
			SUM(CASE WHEN ep.participant_id = $4 THEN ep.share_amount ELSE 0 END)
		FROM expense_participants ep
		JOIN expenses e ON e.id = ep.expense_id
		WHERE e.group_id = $1
	`, fixture.groupID, fixture.payerParticipant, fixture.guestParticipant, fixture.thirdParticipant).Scan(&firstShare, &secondShare, &thirdShare); err != nil {
		t.Fatalf("query shares: %v", err)
	}
	if firstShare != 33_334 || secondShare != 33_333 || thirdShare != 33_333 {
		t.Fatalf("shares = %d, %d, %d; want 33334, 33333, 33333", firstShare, secondShare, thirdShare)
	}
}

func TestEqualExpenseRejectsDuplicateParticipants(t *testing.T) {
	fixture := newEqualExpenseFixture(t)
	req := fixture.equalReqWithParticipants(100_000, []string{fixture.payerParticipant, fixture.guestParticipant, fixture.guestParticipant})

	_, err := fixture.service.CreateEqualExpense(context.Background(), fixture.ownerID, req)
	if !errors.Is(err, expenseConstants.ErrDuplicateParticipants) {
		t.Fatalf("error = %v, want %v", err, expenseConstants.ErrDuplicateParticipants)
	}
}

func TestEqualExpenseRejectsParticipantFromAnotherGroup(t *testing.T) {
	fixture := newEqualExpenseFixture(t)
	foreignGroupID := uuid.NewString()
	foreignParticipantID := uuid.NewString()
	mustExec(t, context.Background(), `INSERT INTO groups (id, name, created_by) VALUES ($1, 'Foreign Group', $2)`, foreignGroupID, fixture.ownerID)
	mustExec(t, context.Background(), `
		INSERT INTO group_participants (id, group_id, user_id, display_name, participant_type, created_by)
		VALUES ($1, $2, NULL, 'Foreign', 'guest', $3)
	`, foreignParticipantID, foreignGroupID, fixture.ownerID)
	t.Cleanup(func() {
		mustExec(t, context.Background(), `DELETE FROM group_participants WHERE id = $1`, foreignParticipantID)
		mustExec(t, context.Background(), `DELETE FROM groups WHERE id = $1`, foreignGroupID)
	})

	req := fixture.equalReqWithParticipants(100_000, []string{fixture.payerParticipant, foreignParticipantID})
	_, err := fixture.service.CreateEqualExpense(context.Background(), fixture.ownerID, req)
	if !errors.Is(err, expenseConstants.ErrParticipantNotInGroup) {
		t.Fatalf("error = %v, want %v", err, expenseConstants.ErrParticipantNotInGroup)
	}
}
