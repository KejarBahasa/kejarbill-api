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
	"github.com/google/uuid"
)

type equalExpenseFixture struct {
	service *expenseServicePkg.ExpenseService

	groupID          string
	ownerID          string
	payerParticipant string
	guestParticipant string
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
		),
		groupID:          uuid.NewString(),
		ownerID:          uuid.NewString(),
		payerParticipant: uuid.NewString(),
		guestParticipant: uuid.NewString(),
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
	return &expenseDto.CreateExpenseEqualRequest{
		GroupID:            f.groupID,
		Title:              "Equal test",
		Currency:           "IDR",
		PayerParticipantID: f.payerParticipant,
		ParticipantIDs:     []string{f.payerParticipant, f.guestParticipant},
		TotalAmount:        totalAmount,
		ExpenseDate:        "2026-08-13T00:00:00Z",
	}
}

func (f *equalExpenseFixture) expenseCount(t *testing.T) int {
	t.Helper()
	return countRows(t, context.Background(), `SELECT COUNT(*) FROM expenses WHERE group_id = $1`, f.groupID)
}

func TestEqualExpenseRejectsNonDivisible(t *testing.T) {
	fixture := newEqualExpenseFixture(t)

	_, err := fixture.service.CreateEqualExpense(context.Background(), fixture.ownerID, fixture.equalReq(101))
	if !errors.Is(err, expenseConstants.ErrEqualAmountNotDivisible) {
		t.Fatalf("CreateEqualExpense() error = %v, want %v", err, expenseConstants.ErrEqualAmountNotDivisible)
	}
	if n := fixture.expenseCount(t); n != 0 {
		t.Fatalf("expense count = %d, want 0", n)
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
