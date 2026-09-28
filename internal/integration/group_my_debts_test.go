package integration_test

import (
	"context"
	"errors"
	"testing"
	"time"

	expenseConstants "github.com/KejarBahasa/kejarbill-api/internal/module/expense/constants"
	groupConstants "github.com/KejarBahasa/kejarbill-api/internal/module/group/constants"
	ledgerRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/ledger/repository"
	"github.com/google/uuid"
)

func (f *summaryFixture) addDebtExpense(t *testing.T, expenseID, title, payerID, expenseDate string, shareAmount int64, deleted bool) {
	t.Helper()
	deletedAt := any(nil)
	if deleted {
		deletedAt = time.Now()
	}
	mustExec(t, context.Background(), `
		INSERT INTO expenses (id, group_id, title, paid_by_participant_id, total_amount, split_method, expense_date, created_by, deleted_at)
		VALUES ($1, $2, $3, $4, $5, 'custom', $6, $7, $8)
	`, expenseID, f.groupID, title, payerID, shareAmount, expenseDate, f.ownerUser, deletedAt)
	mustExec(t, context.Background(), `
		INSERT INTO expense_participants (expense_id, participant_id, share_amount)
		VALUES ($1, $2, $3)
	`, expenseID, f.memberP, shareAmount)
	if !deleted && payerID != f.memberP {
		mustExec(t, context.Background(), `
			INSERT INTO account_ledger (group_id, from_participant_id, to_participant_id, source_type, source_id, amount)
			VALUES ($1, $2, $3, 'expense', $4, $5)
		`, f.groupID, f.memberP, payerID, expenseID, shareAmount)
	}
}

func (f *summaryFixture) addCompletedSettlement(t *testing.T, from, to string, amount int64, paidAt string) {
	t.Helper()
	settlementID := uuid.NewString()
	mustExec(t, context.Background(), `
		INSERT INTO settlements (id, group_id, from_participant_id, to_participant_id, amount, paid_at, status, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, 'completed', $7)
	`, settlementID, f.groupID, from, to, amount, paidAt, f.ownerUser)
	mustExec(t, context.Background(), `
		INSERT INTO account_ledger (group_id, from_participant_id, to_participant_id, source_type, source_id, amount)
		VALUES ($1, $2, $3, 'settlement', $4, $5)
	`, f.groupID, to, from, settlementID, amount)
}

func TestGroupMyDebtsUnpaidAndFIFOAllocation(t *testing.T) {
	f := newSummaryFixture(t)
	firstID := uuid.NewString()
	secondID := uuid.NewString()
	f.addDebtExpense(t, firstID, "Old Expense", f.ownerP, "2026-09-01T12:00:00Z", 50_000, false)
	f.addDebtExpense(t, secondID, "New Expense", f.ownerP, "2026-09-02T12:00:00Z", 40_000, false)
	f.addCompletedSettlement(t, f.memberP, f.ownerP, 20_000, "2026-09-03T12:00:00Z")

	debts, err := f.service.GetMyDebts(context.Background(), f.memberUser, f.groupID)
	if err != nil {
		t.Fatalf("GetMyDebts() error = %v", err)
	}
	if len(debts.Debts) != 1 {
		t.Fatalf("debts length = %d, want 1", len(debts.Debts))
	}
	debt := debts.Debts[0]
	if debt.TotalAmount != 90_000 || debt.PaidAmount != 20_000 || debt.RemainingAmount != 70_000 || debt.Status != groupConstants.DebtStatusPartial {
		t.Fatalf("debt totals = %+v, want total=90000 paid=20000 remaining=70000", debt)
	}
	if len(debt.Expenses) != 2 {
		t.Fatalf("expenses length = %d, want 2", len(debt.Expenses))
	}
	if debt.Expenses[0].ExpenseID != firstID || debt.Expenses[0].Status != groupConstants.DebtStatusPartial || debt.Expenses[0].PaidAmount != 20_000 {
		t.Fatalf("old expense = %+v, want partial paid=20000", debt.Expenses[0])
	}
	if debt.Expenses[1].ExpenseID != secondID || debt.Expenses[1].Status != groupConstants.DebtStatusUnpaid {
		t.Fatalf("new expense = %+v, want unpaid", debt.Expenses[1])
	}

	balances, err := ledgerRepoPkg.NewLedgerRepository().GetGroupBalances(context.Background(), integrationPool, f.groupID)
	if err != nil {
		t.Fatalf("GetGroupBalances() error = %v", err)
	}
	if len(balances) != 1 || balances[0].FromParticipant.ID != f.memberP || balances[0].ToParticipant.ID != f.ownerP || balances[0].Amount != debt.RemainingAmount {
		t.Fatalf("balance = %+v, want member -> owner 70000", balances)
	}
}

func TestGroupMyDebtsEmpty(t *testing.T) {
	f := newSummaryFixture(t)

	debts, err := f.service.GetMyDebts(context.Background(), f.memberUser, f.groupID)
	if err != nil {
		t.Fatalf("GetMyDebts() error = %v", err)
	}
	if len(debts.Debts) != 0 || debts.TotalAmount != 0 || debts.PaidAmount != 0 || debts.RemainingAmount != 0 {
		t.Fatalf("expected zero totals, got %+v", debts)
	}
}

func TestGroupMyDebtsOmitsFullyPaidPair(t *testing.T) {
	f := newSummaryFixture(t)
	expenseID := uuid.NewString()
	f.addDebtExpense(t, expenseID, "Dinner", f.ownerP, "2026-09-01T12:00:00Z", 50_000, false)
	f.addCompletedSettlement(t, f.memberP, f.ownerP, 50_000, "2026-09-02T12:00:00Z")

	debts, err := f.service.GetMyDebts(context.Background(), f.memberUser, f.groupID)
	if err != nil {
		t.Fatalf("GetMyDebts() error = %v", err)
	}
	if len(debts.Debts) != 0 || debts.TotalAmount != 0 || debts.PaidAmount != 0 || debts.RemainingAmount != 0 {
		t.Fatalf("expected empty debt response, got %+v", debts)
	}
}

func TestGroupMyDebtsIgnoresOwnDeletedAndUnrelatedExpenses(t *testing.T) {
	f := newSummaryFixture(t)
	f.addDebtExpense(t, uuid.NewString(), "Own Expense", f.memberP, "2026-09-01T12:00:00Z", 30_000, false)
	f.addDebtExpense(t, uuid.NewString(), "Deleted Expense", f.ownerP, "2026-09-02T12:00:00Z", 40_000, true)
	// This reverse ledger belongs to the member/guest pair, not member/owner.
	f.addCompletedSettlement(t, f.memberP, f.guestP, 10_000, "2026-09-03T12:00:00Z")

	debts, err := f.service.GetMyDebts(context.Background(), f.memberUser, f.groupID)
	if err != nil {
		t.Fatalf("GetMyDebts() error = %v", err)
	}
	if len(debts.Debts) != 0 {
		t.Fatalf("expected no debts, got %+v", debts)
	}
}

func TestGroupMyDebtsRejectsNonMember(t *testing.T) {
	f := newSummaryFixture(t)

	_, err := f.service.GetMyDebts(context.Background(), f.outsider, f.groupID)
	if !errors.Is(err, expenseConstants.ErrForbiddenGroupAccess) {
		t.Fatalf("error = %v, want %v", err, expenseConstants.ErrForbiddenGroupAccess)
	}
}
