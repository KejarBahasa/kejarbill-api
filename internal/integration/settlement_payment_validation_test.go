package integration_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	groupRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/group/repository"
	groupMemberRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/group_member/repository"
	groupParticipantRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/group_participant/repository"
	ledgerConstants "github.com/KejarBahasa/kejarbill-api/internal/module/ledger/constants"
	ledgerRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/ledger/repository"
	paymentMethodConstants "github.com/KejarBahasa/kejarbill-api/internal/module/payment_method/constants"
	paymentMethodRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/payment_method/repository"
	settlementConstants "github.com/KejarBahasa/kejarbill-api/internal/module/settlement/constants"
	settlementDTO "github.com/KejarBahasa/kejarbill-api/internal/module/settlement/dto"
	settlementRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/settlement/repository"
	settlementServicePkg "github.com/KejarBahasa/kejarbill-api/internal/module/settlement/service"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

var integrationPool *pgxpool.Pool

func TestMain(m *testing.M) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		fmt.Fprintln(os.Stderr, "TEST_DATABASE_URL is required for PostgreSQL integration tests")
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect TEST_DATABASE_URL: %v\n", err)
		os.Exit(1)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		fmt.Fprintf(os.Stderr, "ping TEST_DATABASE_URL: %v\n", err)
		os.Exit(1)
	}

	integrationPool = pool
	exitCode := m.Run()
	pool.Close()
	os.Exit(exitCode)
}

func TestSettlementServiceCreatePaymentValidation(t *testing.T) {
	tests := []struct {
		name           string
		paymentChannel string
		paymentType    string
		paymentStatus  string
		methodOwner    paymentMethodOwner
		recipientGuest bool
		useMethod      bool
		unknownMethod  bool
		wantErr        error
	}{
		{
			name:           "cash without payment method succeeds",
			paymentChannel: settlementConstants.PaymentChannelCash,
		},
		{
			name:           "bank transfer with active recipient bank method succeeds",
			paymentChannel: settlementConstants.PaymentChannelBankTransfer,
			paymentType:    paymentMethodConstants.MethodTypeBank,
			paymentStatus:  paymentMethodConstants.StatusActive,
			methodOwner:    paymentMethodRecipient,
			useMethod:      true,
		},
		{
			name:           "ewallet with active recipient ewallet method succeeds",
			paymentChannel: settlementConstants.PaymentChannelEwallet,
			paymentType:    paymentMethodConstants.MethodTypeEWallet,
			paymentStatus:  paymentMethodConstants.StatusActive,
			methodOwner:    paymentMethodRecipient,
			useMethod:      true,
		},
		{
			name:           "cash with payment method fails",
			paymentChannel: settlementConstants.PaymentChannelCash,
			paymentType:    paymentMethodConstants.MethodTypeBank,
			paymentStatus:  paymentMethodConstants.StatusActive,
			methodOwner:    paymentMethodRecipient,
			useMethod:      true,
			wantErr:        settlementConstants.ErrPaymentMethodMustBeEmpty,
		},
		{
			name:           "bank transfer without payment method fails",
			paymentChannel: settlementConstants.PaymentChannelBankTransfer,
			wantErr:        settlementConstants.ErrPaymentMethodRequired,
		},
		{
			name:           "ewallet without payment method fails",
			paymentChannel: settlementConstants.PaymentChannelEwallet,
			wantErr:        settlementConstants.ErrPaymentMethodRequired,
		},
		{
			name:           "bank transfer with ewallet method fails",
			paymentChannel: settlementConstants.PaymentChannelBankTransfer,
			paymentType:    paymentMethodConstants.MethodTypeEWallet,
			paymentStatus:  paymentMethodConstants.StatusActive,
			methodOwner:    paymentMethodRecipient,
			useMethod:      true,
			wantErr:        settlementConstants.ErrInvalidPaymentMethodType,
		},
		{
			name:           "ewallet with bank method fails",
			paymentChannel: settlementConstants.PaymentChannelEwallet,
			paymentType:    paymentMethodConstants.MethodTypeBank,
			paymentStatus:  paymentMethodConstants.StatusActive,
			methodOwner:    paymentMethodRecipient,
			useMethod:      true,
			wantErr:        settlementConstants.ErrInvalidPaymentMethodType,
		},
		{
			name:           "hidden payment method fails",
			paymentChannel: settlementConstants.PaymentChannelBankTransfer,
			paymentType:    paymentMethodConstants.MethodTypeBank,
			paymentStatus:  paymentMethodConstants.StatusHidden,
			methodOwner:    paymentMethodRecipient,
			useMethod:      true,
			wantErr:        paymentMethodConstants.ErrPaymentMethodInactive,
		},
		{
			name:           "deleted payment method fails",
			paymentChannel: settlementConstants.PaymentChannelBankTransfer,
			paymentType:    paymentMethodConstants.MethodTypeBank,
			paymentStatus:  paymentMethodConstants.StatusDeleted,
			methodOwner:    paymentMethodRecipient,
			useMethod:      true,
			wantErr:        paymentMethodConstants.ErrPaymentMethodInactive,
		},
		{
			name:           "payment method owned by another user fails",
			paymentChannel: settlementConstants.PaymentChannelBankTransfer,
			paymentType:    paymentMethodConstants.MethodTypeBank,
			paymentStatus:  paymentMethodConstants.StatusActive,
			methodOwner:    paymentMethodOtherUser,
			useMethod:      true,
			wantErr:        settlementConstants.ErrPaymentMethodNotOwned,
		},
		{
			name:           "guest recipient with payment method fails",
			paymentChannel: settlementConstants.PaymentChannelBankTransfer,
			paymentType:    paymentMethodConstants.MethodTypeBank,
			paymentStatus:  paymentMethodConstants.StatusActive,
			methodOwner:    paymentMethodOtherUser,
			recipientGuest: true,
			useMethod:      true,
			wantErr:        settlementConstants.ErrPaymentMethodNotOwned,
		},
		{
			name:           "unknown payment method fails",
			paymentChannel: settlementConstants.PaymentChannelBankTransfer,
			unknownMethod:  true,
			wantErr:        paymentMethodConstants.ErrPaymentMethodNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := newSettlementFixture(t, test.recipientGuest)
			paymentMethodID := ""
			if test.useMethod {
				ownerID := fixture.recipientUserID
				if test.methodOwner == paymentMethodOtherUser {
					ownerID = fixture.otherUserID
				}
				paymentMethodID = fixture.createPaymentMethod(t, ownerID, test.paymentType, test.paymentStatus)
			}
			if test.unknownMethod {
				paymentMethodID = uuid.NewString()
			}

			request := &settlementDTO.CreateSettlementBody{
				FromParticipantID: fixture.fromParticipantID,
				ToParticipantID:   fixture.toParticipantID,
				Amount:            40_000,
				PaymentChannel:    test.paymentChannel,
				PaymentMethodID:   paymentMethodID,
				Notes:             "integration validation test",
				PaidAt:            time.Now().UTC().Format(time.RFC3339),
			}
			idempotencyKey := uuid.NewString()

			result, err := fixture.service.Create(context.Background(), fixture.requesterUserID, fixture.groupID, idempotencyKey, request)
			if test.wantErr == nil {
				if err != nil {
					t.Fatalf("Create() error = %v", err)
				}
				if result == nil || result.ResponseCode != 200 || result.ResponseBody == "" {
					t.Fatalf("Create() result = %#v, want stored success response", result)
				}
				fixture.assertCommittedSettlement(t, idempotencyKey)
				return
			}

			if !errors.Is(err, test.wantErr) {
				t.Fatalf("Create() error = %v, want %v", err, test.wantErr)
			}
			fixture.assertNoCommittedSettlement(t, idempotencyKey)
		})
	}
}

type paymentMethodOwner int

const (
	paymentMethodRecipient paymentMethodOwner = iota
	paymentMethodOtherUser
)

type settlementFixture struct {
	service *settlementServicePkg.SettlementService

	groupID           string
	requesterUserID   string
	recipientUserID   string
	otherUserID       string
	fromParticipantID string
	toParticipantID   string
	expenseID         string
}

func newSettlementFixture(t *testing.T, recipientGuest bool) *settlementFixture {
	t.Helper()

	fixture := &settlementFixture{
		groupID:           uuid.NewString(),
		requesterUserID:   uuid.NewString(),
		recipientUserID:   uuid.NewString(),
		otherUserID:       uuid.NewString(),
		fromParticipantID: uuid.NewString(),
		toParticipantID:   uuid.NewString(),
		expenseID:         uuid.NewString(),
	}
	fixture.service = settlementServicePkg.NewSettlementService(
		integrationPool,
		settlementRepoPkg.NewSettlementRepository(),
		ledgerRepoPkg.NewLedgerRepository(),
		groupRepoPkg.NewGroupRepository(),
		groupMemberRepoPkg.NewGroupMemberRepository(),
		groupParticipantRepoPkg.NewGroupParticipantRepository(),
		paymentMethodRepoPkg.NewPaymentMethodRepository(),
	)

	ctx := context.Background()
	for _, userID := range []string{fixture.requesterUserID, fixture.recipientUserID, fixture.otherUserID} {
		mustExec(t, ctx, `
			INSERT INTO users (id, email, username, full_name, password_hash)
			VALUES ($1, $2, $3, 'Integration Test User', 'not-a-real-password')
		`, userID, userID+"@example.test", "user_"+userID[:8])
	}

	mustExec(t, ctx, `
		INSERT INTO groups (id, name, created_by)
		VALUES ($1, 'Settlement Integration Test', $2)
	`, fixture.groupID, fixture.requesterUserID)
	mustExec(t, ctx, `
		INSERT INTO group_members (group_id, user_id, role, status)
		VALUES ($1, $2, 'owner', 'active')
	`, fixture.groupID, fixture.requesterUserID)
	mustExec(t, ctx, `
		INSERT INTO group_participants (id, group_id, user_id, display_name, participant_type, created_by)
		VALUES ($1, $2, $3, 'Debtor', 'registered', $3)
	`, fixture.fromParticipantID, fixture.groupID, fixture.requesterUserID)

	if recipientGuest {
		mustExec(t, ctx, `
			INSERT INTO group_participants (id, group_id, display_name, participant_type, created_by)
			VALUES ($1, $2, 'Guest Recipient', 'guest', $3)
		`, fixture.toParticipantID, fixture.groupID, fixture.requesterUserID)
	} else {
		mustExec(t, ctx, `
			INSERT INTO group_participants (id, group_id, user_id, display_name, participant_type, created_by)
			VALUES ($1, $2, $3, 'Recipient', 'registered', $4)
		`, fixture.toParticipantID, fixture.groupID, fixture.recipientUserID, fixture.requesterUserID)
	}

	mustExec(t, ctx, `
		INSERT INTO expenses (
			id, group_id, title, paid_by_participant_id, total_amount, split_method, expense_date, created_by
		)
		VALUES ($1, $2, 'Fixture obligation', $3, 100000, 'custom', NOW(), $4)
	`, fixture.expenseID, fixture.groupID, fixture.toParticipantID, fixture.requesterUserID)
	mustExec(t, ctx, `
		INSERT INTO account_ledger (
			group_id, from_participant_id, to_participant_id, source_type, source_id, amount
		)
		VALUES ($1, $2, $3, $4, $5, 100000)
	`, fixture.groupID, fixture.fromParticipantID, fixture.toParticipantID, ledgerConstants.SourceTypeExpense, fixture.expenseID)

	t.Cleanup(func() { fixture.cleanup(t) })
	return fixture
}

func (f *settlementFixture) createPaymentMethod(t *testing.T, userID, methodType, status string) string {
	t.Helper()

	paymentMethodID := uuid.NewString()
	mustExec(t, context.Background(), `
		INSERT INTO payment_methods (id, user_id, method_type, provider_name, status)
		VALUES ($1, $2, $3, 'Integration Test Provider', $4)
	`, paymentMethodID, userID, methodType, status)
	return paymentMethodID
}

func (f *settlementFixture) assertCommittedSettlement(t *testing.T, idempotencyKey string) {
	t.Helper()
	ctx := context.Background()

	if countRows(t, ctx, `SELECT COUNT(*) FROM settlements WHERE group_id = $1`, f.groupID) != 1 {
		t.Fatal("expected one committed settlement")
	}
	if countRows(t, ctx, `SELECT COUNT(*) FROM account_ledger WHERE group_id = $1 AND source_type = $2`, f.groupID, ledgerConstants.SourceTypeSettlement) != 1 {
		t.Fatal("expected one committed settlement ledger row")
	}
	if countRows(t, ctx, `SELECT COUNT(*) FROM idempotency_keys WHERE user_id = $1 AND idempotency_key = $2`, f.requesterUserID, idempotencyKey) != 1 {
		t.Fatal("expected one committed idempotency record")
	}
}

func (f *settlementFixture) assertNoCommittedSettlement(t *testing.T, idempotencyKey string) {
	t.Helper()
	ctx := context.Background()

	if countRows(t, ctx, `SELECT COUNT(*) FROM settlements WHERE group_id = $1`, f.groupID) != 0 {
		t.Fatal("expected no committed settlement after validation failure")
	}
	if countRows(t, ctx, `SELECT COUNT(*) FROM account_ledger WHERE group_id = $1 AND source_type = $2`, f.groupID, ledgerConstants.SourceTypeSettlement) != 0 {
		t.Fatal("expected no committed settlement ledger row after validation failure")
	}
	if countRows(t, ctx, `SELECT COUNT(*) FROM idempotency_keys WHERE user_id = $1 AND idempotency_key = $2`, f.requesterUserID, idempotencyKey) != 0 {
		t.Fatal("expected no committed idempotency record after validation failure")
	}
}

func (f *settlementFixture) cleanup(t *testing.T) {
	t.Helper()
	ctx := context.Background()

	for _, statement := range []struct {
		query string
		args  []any
	}{
		{`DELETE FROM account_ledger WHERE group_id = $1`, []any{f.groupID}},
		{`DELETE FROM settlements WHERE group_id = $1`, []any{f.groupID}},
		{`DELETE FROM expense_participants WHERE expense_id = $1`, []any{f.expenseID}},
		{`DELETE FROM expenses WHERE id = $1`, []any{f.expenseID}},
		{`DELETE FROM idempotency_keys WHERE user_id = ANY($1)`, []any{[]string{f.requesterUserID, f.recipientUserID, f.otherUserID}}},
		{`DELETE FROM payment_methods WHERE user_id = ANY($1)`, []any{[]string{f.requesterUserID, f.recipientUserID, f.otherUserID}}},
		{`DELETE FROM group_participants WHERE group_id = $1`, []any{f.groupID}},
		{`DELETE FROM group_members WHERE group_id = $1`, []any{f.groupID}},
		{`DELETE FROM groups WHERE id = $1`, []any{f.groupID}},
		{`DELETE FROM users WHERE id = ANY($1)`, []any{[]string{f.requesterUserID, f.recipientUserID, f.otherUserID}}},
	} {
		if _, err := integrationPool.Exec(ctx, statement.query, statement.args...); err != nil {
			t.Errorf("cleanup query failed: %v", err)
		}
	}
}

func mustExec(t *testing.T, ctx context.Context, query string, args ...any) {
	t.Helper()
	if _, err := integrationPool.Exec(ctx, query, args...); err != nil {
		t.Fatalf("fixture query failed: %v", err)
	}
}

func countRows(t *testing.T, ctx context.Context, query string, args ...any) int {
	t.Helper()

	var count int
	if err := integrationPool.QueryRow(ctx, query, args...).Scan(&count); err != nil {
		t.Fatalf("count query failed: %v", err)
	}
	return count
}
