package integration_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	groupRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/group/repository"
	groupMemberRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/group_member/repository"
	groupParticipantRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/group_participant/repository"
	ledgerConstants "github.com/KejarBahasa/kejarbill-api/internal/module/ledger/constants"
	ledgerDTO "github.com/KejarBahasa/kejarbill-api/internal/module/ledger/dto"
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
	databaseURL := "postgres://mramdhani:@localhost:5432/kejarbill-test?sslmode=disable"
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

func TestSettlementServiceCreateIdempotencySameKeyIdenticalPayload(t *testing.T) {
	fixture := newSettlementFixture(t, false)
	request := fixture.cashSettlementRequest(40_000)
	idempotencyKey := "test-idempotency-1"

	first, err := fixture.service.Create(context.Background(), fixture.requesterUserID, fixture.groupID, idempotencyKey, request)
	if err != nil {
		t.Fatalf("first Create() error = %v", err)
	}
	second, err := fixture.service.Create(context.Background(), fixture.requesterUserID, fixture.groupID, idempotencyKey, request)
	if err != nil {
		t.Fatalf("second Create() error = %v", err)
	}

	firstSettlementID := settlementIDFromResponse(t, first.ResponseBody)
	secondSettlementID := settlementIDFromResponse(t, second.ResponseBody)
	if firstSettlementID != secondSettlementID {
		t.Fatalf("settlement IDs differ: first = %q, second = %q", firstSettlementID, secondSettlementID)
	}

	storedResponseCode, storedResponseBody := fixture.idempotencyResponse(t, idempotencyKey)
	if second.ResponseCode != storedResponseCode || second.ResponseBody != storedResponseBody {
		t.Fatalf("second result = (%d, %q), want stored response = (%d, %q)", second.ResponseCode, second.ResponseBody, storedResponseCode, storedResponseBody)
	}
	fixture.assertCommittedSettlement(t, idempotencyKey)
}

func TestSettlementServiceCreateIdempotencySameKeyDifferentPayload(t *testing.T) {
	fixture := newSettlementFixture(t, false)
	idempotencyKey := "test-idempotency-2"

	first, err := fixture.service.Create(context.Background(), fixture.requesterUserID, fixture.groupID, idempotencyKey, fixture.cashSettlementRequest(40_000))
	if err != nil {
		t.Fatalf("first Create() error = %v", err)
	}
	if first == nil {
		t.Fatal("first Create() result is nil")
	}

	_, err = fixture.service.Create(context.Background(), fixture.requesterUserID, fixture.groupID, idempotencyKey, fixture.cashSettlementRequest(50_000))
	if !errors.Is(err, settlementConstants.ErrIdempotencyKeyConflict) {
		t.Fatalf("second Create() error = %v, want %v", err, settlementConstants.ErrIdempotencyKeyConflict)
	}
	fixture.assertCommittedSettlement(t, idempotencyKey)
}

func TestSettlementServiceCreateIdempotencyRollbackThenRetry(t *testing.T) {
	fixture := newSettlementFixture(t, false)
	idempotencyKey := "test-idempotency-rollback-retry"
	request := fixture.cashSettlementRequest(40_000)

	mustExec(t, context.Background(), `
		DELETE FROM account_ledger
		WHERE group_id = $1 AND source_type = $2 AND source_id = $3
	`, fixture.groupID, ledgerConstants.SourceTypeExpense, fixture.expenseID)

	_, err := fixture.service.Create(context.Background(), fixture.requesterUserID, fixture.groupID, idempotencyKey, request)
	if !errors.Is(err, settlementConstants.ErrSettlementAmountExceeded) {
		t.Fatalf("first Create() error = %v, want %v", err, settlementConstants.ErrSettlementAmountExceeded)
	}
	fixture.assertNoCommittedSettlement(t, idempotencyKey)

	mustExec(t, context.Background(), `
		INSERT INTO account_ledger (
			group_id, from_participant_id, to_participant_id, source_type, source_id, amount
		)
		VALUES ($1, $2, $3, $4, $5, 100000)
	`, fixture.groupID, fixture.fromParticipantID, fixture.toParticipantID, ledgerConstants.SourceTypeExpense, fixture.expenseID)

	result, err := fixture.service.Create(context.Background(), fixture.requesterUserID, fixture.groupID, idempotencyKey, request)
	if err != nil {
		t.Fatalf("retry Create() error = %v", err)
	}
	if result == nil || result.ResponseCode != 200 || settlementIDFromResponse(t, result.ResponseBody) == "" {
		t.Fatalf("retry Create() result = %#v, want successful settlement response", result)
	}
	fixture.assertCommittedSettlement(t, idempotencyKey)
}

func TestSettlementServiceCreateConcurrentFullSettlements(t *testing.T) {
	fixture := newSettlementFixture(t, false)
	outcomes := runConcurrentSettlementCreates(
		t,
		fixture,
		"test-concurrent-full-"+uuid.NewString(), fixture.cashSettlementRequest(100_000),
		"test-concurrent-full-"+uuid.NewString(), fixture.cashSettlementRequest(100_000),
	)

	assertOneSuccessAndOneAmountExceeded(t, outcomes)
	fixture.assertSettlementCounts(t, 1, 1)
	if outstanding := fixture.outstandingBalance(t); outstanding != 0 {
		t.Fatalf("outstanding balance = %d, want 0", outstanding)
	}
}

func TestSettlementServiceCreateConcurrentPartialSettlements(t *testing.T) {
	fixture := newSettlementFixture(t, false)
	outcomes := runConcurrentSettlementCreates(
		t,
		fixture,
		"test-concurrent-partial-"+uuid.NewString(), fixture.cashSettlementRequest(60_000),
		"test-concurrent-partial-"+uuid.NewString(), fixture.cashSettlementRequest(60_000),
	)

	assertOneSuccessAndOneAmountExceeded(t, outcomes)
	fixture.assertSettlementCounts(t, 1, 1)
	if outstanding := fixture.outstandingBalance(t); outstanding != 40_000 {
		t.Fatalf("outstanding balance = %d, want 40000", outstanding)
	}
}

func TestSettlementServiceCreateConcurrentSameIdempotencyKey(t *testing.T) {
	fixture := newSettlementFixture(t, false)
	idempotencyKey := "test-concurrent-same-key-" + uuid.NewString()
	outcomes := runConcurrentSettlementCreates(
		t,
		fixture,
		idempotencyKey, fixture.cashSettlementRequest(40_000),
		idempotencyKey, fixture.cashSettlementRequest(40_000),
	)

	if outcomes[0].err != nil || outcomes[1].err != nil {
		t.Fatalf("concurrent Create() errors = [%v, %v], want both successful", outcomes[0].err, outcomes[1].err)
	}
	if outcomes[0].result == nil || outcomes[1].result == nil {
		t.Fatalf("concurrent Create() results = %#v, %#v, want successful results", outcomes[0].result, outcomes[1].result)
	}
	firstSettlementID := settlementIDFromResponse(t, outcomes[0].result.ResponseBody)
	secondSettlementID := settlementIDFromResponse(t, outcomes[1].result.ResponseBody)
	if firstSettlementID != secondSettlementID {
		t.Fatalf("settlement IDs differ: %q and %q", firstSettlementID, secondSettlementID)
	}

	fixture.assertSettlementCounts(t, 1, 1)
	if countRows(t, context.Background(), `SELECT COUNT(*) FROM idempotency_keys WHERE idempotency_key = $1`, idempotencyKey) != 1 {
		t.Fatal("expected exactly one committed idempotency record")
	}
	if outstanding := fixture.outstandingBalance(t); outstanding != 60_000 {
		t.Fatalf("outstanding balance = %d, want 60000", outstanding)
	}
}

func TestLedgerRepositoryGetGroupBalancesReciprocalNetting(t *testing.T) {
	tests := []struct {
		name     string
		entries  func(t *testing.T, fixture *settlementFixture) []ledgerEntry
		expected func(fixture *settlementFixture, participantC string) []expectedBalance
	}{
		{
			name: "single obligation",
			entries: func(_ *testing.T, fixture *settlementFixture) []ledgerEntry {
				return []ledgerEntry{{fixture.fromParticipantID, fixture.toParticipantID, 100_000}}
			},
			expected: func(fixture *settlementFixture, _ string) []expectedBalance {
				return []expectedBalance{{fixture.fromParticipantID, fixture.toParticipantID, 100_000}}
			},
		},
		{
			name: "partial reciprocal settlement",
			entries: func(_ *testing.T, fixture *settlementFixture) []ledgerEntry {
				return []ledgerEntry{
					{fixture.fromParticipantID, fixture.toParticipantID, 100_000},
					{fixture.toParticipantID, fixture.fromParticipantID, 40_000},
				}
			},
			expected: func(fixture *settlementFixture, _ string) []expectedBalance {
				return []expectedBalance{{fixture.fromParticipantID, fixture.toParticipantID, 60_000}}
			},
		},
		{
			name: "full reciprocal settlement omits net zero pair",
			entries: func(_ *testing.T, fixture *settlementFixture) []ledgerEntry {
				return []ledgerEntry{
					{fixture.fromParticipantID, fixture.toParticipantID, 100_000},
					{fixture.toParticipantID, fixture.fromParticipantID, 100_000},
				}
			},
			expected: func(_ *settlementFixture, _ string) []expectedBalance {
				return nil
			},
		},
		{
			name: "reversed net direction",
			entries: func(_ *testing.T, fixture *settlementFixture) []ledgerEntry {
				return []ledgerEntry{
					{fixture.fromParticipantID, fixture.toParticipantID, 40_000},
					{fixture.toParticipantID, fixture.fromParticipantID, 100_000},
				}
			},
			expected: func(fixture *settlementFixture, _ string) []expectedBalance {
				return []expectedBalance{{fixture.toParticipantID, fixture.fromParticipantID, 60_000}}
			},
		},
		{
			name: "multiple rows in both directions",
			entries: func(_ *testing.T, fixture *settlementFixture) []ledgerEntry {
				return []ledgerEntry{
					{fixture.fromParticipantID, fixture.toParticipantID, 100_000},
					{fixture.fromParticipantID, fixture.toParticipantID, 50_000},
					{fixture.toParticipantID, fixture.fromParticipantID, 20_000},
					{fixture.toParticipantID, fixture.fromParticipantID, 10_000},
				}
			},
			expected: func(fixture *settlementFixture, _ string) []expectedBalance {
				return []expectedBalance{{fixture.fromParticipantID, fixture.toParticipantID, 120_000}}
			},
		},
		{
			name: "multiple participant pairs remain independent",
			entries: func(t *testing.T, fixture *settlementFixture) []ledgerEntry {
				participantC := fixture.createGuestParticipant(t, "Participant C")
				return []ledgerEntry{
					{fixture.fromParticipantID, fixture.toParticipantID, 100_000},
					{fixture.fromParticipantID, participantC, 50_000},
					{fixture.toParticipantID, participantC, 30_000},
				}
			},
			expected: func(fixture *settlementFixture, participantC string) []expectedBalance {
				return []expectedBalance{
					{fixture.fromParticipantID, fixture.toParticipantID, 100_000},
					{fixture.fromParticipantID, participantC, 50_000},
					{fixture.toParticipantID, participantC, 30_000},
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := newSettlementFixture(t, false)
			mustExec(t, context.Background(), `DELETE FROM account_ledger WHERE group_id = $1`, fixture.groupID)

			entries := test.entries(t, fixture)
			participantC := ""
			for _, entry := range entries {
				if entry.fromParticipantID != fixture.fromParticipantID && entry.fromParticipantID != fixture.toParticipantID {
					participantC = entry.fromParticipantID
				}
				if entry.toParticipantID != fixture.fromParticipantID && entry.toParticipantID != fixture.toParticipantID {
					participantC = entry.toParticipantID
				}
				fixture.insertLedgerEntry(t, entry)
			}

			balances, err := ledgerRepoPkg.NewLedgerRepository().GetGroupBalances(context.Background(), integrationPool, fixture.groupID)
			if err != nil {
				t.Fatalf("GetGroupBalances() error = %v", err)
			}
			assertBalances(t, balances, test.expected(fixture, participantC))
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

type ledgerEntry struct {
	fromParticipantID string
	toParticipantID   string
	amount            int64
}

type expectedBalance struct {
	fromParticipantID string
	toParticipantID   string
	amount            int64
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

func (f *settlementFixture) createGuestParticipant(t *testing.T, displayName string) string {
	t.Helper()

	participantID := uuid.NewString()
	mustExec(t, context.Background(), `
		INSERT INTO group_participants (id, group_id, display_name, participant_type, created_by)
		VALUES ($1, $2, $3, 'guest', $4)
	`, participantID, f.groupID, displayName, f.requesterUserID)
	return participantID
}

func (f *settlementFixture) insertLedgerEntry(t *testing.T, entry ledgerEntry) {
	t.Helper()

	mustExec(t, context.Background(), `
		INSERT INTO account_ledger (
			group_id, from_participant_id, to_participant_id, source_type, source_id, amount
		)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, f.groupID, entry.fromParticipantID, entry.toParticipantID, ledgerConstants.SourceTypeAdjustment, uuid.NewString(), entry.amount)
}

func (f *settlementFixture) cashSettlementRequest(amount int64) *settlementDTO.CreateSettlementBody {
	return &settlementDTO.CreateSettlementBody{
		FromParticipantID: f.fromParticipantID,
		ToParticipantID:   f.toParticipantID,
		Amount:            amount,
		PaymentChannel:    settlementConstants.PaymentChannelCash,
		Notes:             "integration idempotency test",
		PaidAt:            "2026-08-13T00:00:00Z",
	}
}

func (f *settlementFixture) idempotencyResponse(t *testing.T, idempotencyKey string) (int, string) {
	t.Helper()

	var responseCode int
	var responseBody string
	if err := integrationPool.QueryRow(context.Background(), `
		SELECT response_code, response_body::TEXT
		FROM idempotency_keys
		WHERE user_id = $1 AND idempotency_key = $2
	`, f.requesterUserID, idempotencyKey).Scan(&responseCode, &responseBody); err != nil {
		t.Fatalf("query idempotency response: %v", err)
	}
	return responseCode, responseBody
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

func (f *settlementFixture) assertSettlementCounts(t *testing.T, settlements, settlementLedgerRows int) {
	t.Helper()
	ctx := context.Background()

	if actual := countRows(t, ctx, `SELECT COUNT(*) FROM settlements WHERE group_id = $1`, f.groupID); actual != settlements {
		t.Fatalf("settlement count = %d, want %d", actual, settlements)
	}
	if actual := countRows(t, ctx, `SELECT COUNT(*) FROM account_ledger WHERE group_id = $1 AND source_type = $2`, f.groupID, ledgerConstants.SourceTypeSettlement); actual != settlementLedgerRows {
		t.Fatalf("settlement ledger count = %d, want %d", actual, settlementLedgerRows)
	}
}

func (f *settlementFixture) outstandingBalance(t *testing.T) int64 {
	t.Helper()

	var outstanding int64
	if err := integrationPool.QueryRow(context.Background(), `
		SELECT COALESCE(SUM(
			CASE
				WHEN from_participant_id = $2 AND to_participant_id = $3 THEN amount
				WHEN from_participant_id = $3 AND to_participant_id = $2 THEN -amount
				ELSE 0
			END
		), 0)::BIGINT
		FROM account_ledger
		WHERE group_id = $1
	`, f.groupID, f.fromParticipantID, f.toParticipantID).Scan(&outstanding); err != nil {
		t.Fatalf("query outstanding balance: %v", err)
	}
	return outstanding
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

func settlementIDFromResponse(t *testing.T, responseBody string) string {
	t.Helper()

	var payload struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(responseBody), &payload); err != nil {
		t.Fatalf("unmarshal settlement response: %v", err)
	}
	if payload.Data.ID == "" {
		t.Fatal("settlement response does not contain an ID")
	}
	return payload.Data.ID
}

func assertBalances(t *testing.T, balances []ledgerDTO.GroupBalanceResponse, expected []expectedBalance) {
	t.Helper()

	if len(balances) != len(expected) {
		t.Fatalf("balance count = %d, want %d; actual = %#v", len(balances), len(expected), balances)
	}

	actualByDirection := make(map[string]int64, len(balances))
	actualPairs := make(map[string]struct{}, len(balances))
	for _, balance := range balances {
		direction := balance.FromParticipant.ID + "->" + balance.ToParticipant.ID
		if _, exists := actualByDirection[direction]; exists {
			t.Fatalf("duplicate balance direction %q", direction)
		}
		actualByDirection[direction] = balance.Amount

		pair := unorderedParticipantPair(balance.FromParticipant.ID, balance.ToParticipant.ID)
		if _, exists := actualPairs[pair]; exists {
			t.Fatalf("duplicate balance row for participant pair %q", pair)
		}
		actualPairs[pair] = struct{}{}
	}

	for _, want := range expected {
		direction := want.fromParticipantID + "->" + want.toParticipantID
		if amount, exists := actualByDirection[direction]; !exists || amount != want.amount {
			t.Fatalf("balance %q = %d (exists=%t), want %d; actual = %#v", direction, amount, exists, want.amount, balances)
		}
	}
}

func unorderedParticipantPair(firstParticipantID, secondParticipantID string) string {
	if firstParticipantID < secondParticipantID {
		return firstParticipantID + ":" + secondParticipantID
	}
	return secondParticipantID + ":" + firstParticipantID
}

type settlementCreateOutcome struct {
	result *settlementServicePkg.CreateSettlementResult
	err    error
}

func runConcurrentSettlementCreates(
	t *testing.T,
	fixture *settlementFixture,
	firstKey string,
	firstRequest *settlementDTO.CreateSettlementBody,
	secondKey string,
	secondRequest *settlementDTO.CreateSettlementBody,
) []settlementCreateOutcome {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	start := make(chan struct{})
	ready := make(chan struct{}, 2)
	outcomes := make(chan settlementCreateOutcome, 2)
	var workers sync.WaitGroup

	for _, input := range []struct {
		idempotencyKey string
		request        *settlementDTO.CreateSettlementBody
	}{
		{firstKey, firstRequest},
		{secondKey, secondRequest},
	} {
		workers.Add(1)
		go func(idempotencyKey string, request *settlementDTO.CreateSettlementBody) {
			defer workers.Done()
			ready <- struct{}{}
			<-start
			result, err := fixture.service.Create(ctx, fixture.requesterUserID, fixture.groupID, idempotencyKey, request)
			outcomes <- settlementCreateOutcome{result: result, err: err}
		}(input.idempotencyKey, input.request)
	}

	<-ready
	<-ready
	close(start)
	workers.Wait()
	close(outcomes)

	result := make([]settlementCreateOutcome, 0, 2)
	for outcome := range outcomes {
		result = append(result, outcome)
	}
	return result
}

func assertOneSuccessAndOneAmountExceeded(t *testing.T, outcomes []settlementCreateOutcome) {
	t.Helper()

	if len(outcomes) != 2 {
		t.Fatalf("outcome count = %d, want 2", len(outcomes))
	}

	successes := 0
	amountExceeded := 0
	for _, outcome := range outcomes {
		if outcome.err == nil {
			successes++
			continue
		}
		if errors.Is(outcome.err, settlementConstants.ErrSettlementAmountExceeded) {
			amountExceeded++
			continue
		}
		t.Fatalf("Create() error = %v, want success or %v", outcome.err, settlementConstants.ErrSettlementAmountExceeded)
	}
	if successes != 1 || amountExceeded != 1 {
		t.Fatalf("successes = %d, amount exceeded errors = %d, want one each", successes, amountExceeded)
	}
}
