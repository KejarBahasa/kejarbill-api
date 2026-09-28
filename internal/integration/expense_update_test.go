package integration_test

import (
	"context"
	"errors"
	"testing"

	expenseConstants "github.com/KejarBahasa/kejarbill-api/internal/module/expense/constants"
	expenseDto "github.com/KejarBahasa/kejarbill-api/internal/module/expense/dto"
	ledgerConstants "github.com/KejarBahasa/kejarbill-api/internal/module/ledger/constants"
	"github.com/google/uuid"
)

func updateRequest(f *equalExpenseFixture, total int64) *expenseDto.UpdateExpenseRequest {
	return &expenseDto.UpdateExpenseRequest{
		Title:              "Updated expense",
		Currency:           "IDR",
		ExpenseDate:        "2026-08-14T00:00:00Z",
		PayerParticipantID: f.payerParticipant,
		SplitMethod:        "equal",
		Version:            1,
		TotalAmount:        total,
		ParticipantIDs:     []string{f.payerParticipant, f.guestParticipant},
	}
}

func TestExpenseUpdateBeforeSettlementRebuildsSharesAndLedger(t *testing.T) {
	f := newEqualExpenseFixture(t)
	expenseID, err := f.service.CreateEqualExpense(context.Background(), f.ownerID, f.equalReq(100))
	if err != nil {
		t.Fatalf("CreateEqualExpense() error = %v", err)
	}

	if err := f.service.Update(context.Background(), f.ownerID, expenseID, updateRequest(f, 201)); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	var total, guestShare, ledgerAmount int64
	if err := integrationPool.QueryRow(context.Background(), `SELECT total_amount FROM expenses WHERE id = $1`, expenseID).Scan(&total); err != nil {
		t.Fatalf("query expense: %v", err)
	}
	if err := integrationPool.QueryRow(context.Background(), `
		SELECT share_amount FROM expense_participants WHERE expense_id = $1 AND participant_id = $2
	`, expenseID, f.guestParticipant).Scan(&guestShare); err != nil {
		t.Fatalf("query share: %v", err)
	}
	if err := integrationPool.QueryRow(context.Background(), `
		SELECT amount FROM account_ledger WHERE source_id = $1 AND from_participant_id = $2
	`, expenseID, f.guestParticipant).Scan(&ledgerAmount); err != nil {
		t.Fatalf("query ledger: %v", err)
	}
	if total != 201 || guestShare != 100 || ledgerAmount != 100 {
		t.Fatalf("updated values = total %d share %d ledger %d, want 201/100/100", total, guestShare, ledgerAmount)
	}
	if err := f.service.Update(context.Background(), f.ownerID, expenseID, updateRequest(f, 301)); !errors.Is(err, expenseConstants.ErrExpenseVersionConflict) {
		t.Fatalf("stale Update() error = %v, want version conflict", err)
	}
}

func TestExpenseUpdateAndDeleteRejectNonMember(t *testing.T) {
	f := newEqualExpenseFixture(t)
	expenseID, err := f.service.CreateEqualExpense(context.Background(), f.ownerID, f.equalReq(100))
	if err != nil {
		t.Fatalf("CreateEqualExpense() error = %v", err)
	}
	outsider := uuid.NewString()

	if err := f.service.Update(context.Background(), outsider, expenseID, updateRequest(f, 200)); !errors.Is(err, expenseConstants.ErrForbiddenGroupAccess) {
		t.Fatalf("Update() error = %v, want forbidden group access", err)
	}
	if err := f.service.DeleteByID(context.Background(), outsider, expenseID); !errors.Is(err, expenseConstants.ErrForbiddenGroupAccess) {
		t.Fatalf("DeleteByID() error = %v, want forbidden group access", err)
	}
}

func TestExpenseUpdateAndDeleteLockedAfterSettlement(t *testing.T) {
	f := newEqualExpenseFixture(t)
	expenseID, err := f.service.CreateEqualExpense(context.Background(), f.ownerID, f.equalReq(100))
	if err != nil {
		t.Fatalf("CreateEqualExpense() error = %v", err)
	}
	settlementID := uuid.NewString()
	mustExec(t, context.Background(), `
		INSERT INTO settlements (id, group_id, from_participant_id, to_participant_id, amount, paid_at, status, created_by)
		VALUES ($1, $2, $3, $4, 20, '2026-08-14T00:00:00Z', 'completed', $5)
	`, settlementID, f.groupID, f.guestParticipant, f.payerParticipant, f.ownerID)
	mustExec(t, context.Background(), `
		INSERT INTO account_ledger (group_id, from_participant_id, to_participant_id, source_type, source_id, amount)
		VALUES ($1, $2, $3, $4, $5, 20)
	`, f.groupID, f.payerParticipant, f.guestParticipant, ledgerConstants.SourceTypeSettlement, settlementID)

	if err := f.service.Update(context.Background(), f.ownerID, expenseID, updateRequest(f, 200)); !errors.Is(err, expenseConstants.ErrExpenseLocked) {
		t.Fatalf("Update() error = %v, want expense locked", err)
	}
	if err := f.service.DeleteByID(context.Background(), f.ownerID, expenseID); !errors.Is(err, expenseConstants.ErrExpenseLocked) {
		t.Fatalf("DeleteByID() error = %v, want expense locked", err)
	}
}
