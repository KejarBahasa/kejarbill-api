package integration_test

import (
	"context"
	"errors"
	"testing"

	expenseConstants "github.com/KejarBahasa/kejarbill-api/internal/module/expense/constants"
	groupRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/group/repository"
	groupMemberRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/group_member/repository"
	groupParticipantRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/group_participant/repository"
	ledgerConstants "github.com/KejarBahasa/kejarbill-api/internal/module/ledger/constants"
	ledgerRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/ledger/repository"
	paymentMethodRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/payment_method/repository"
	settlementConstants "github.com/KejarBahasa/kejarbill-api/internal/module/settlement/constants"
	settlementRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/settlement/repository"
	settlementServicePkg "github.com/KejarBahasa/kejarbill-api/internal/module/settlement/service"
	"github.com/google/uuid"
)

type recipientMethodFixture struct {
	service *settlementServicePkg.SettlementService

	groupID              string
	requesterID          string
	requesterParticipant string
	recipientID          string
	recipientParticipant string

	extraUsers []string
}

func newRecipientMethodFixture(t *testing.T) *recipientMethodFixture {
	t.Helper()

	f := &recipientMethodFixture{
		service: settlementServicePkg.NewSettlementService(
			integrationPool,
			settlementRepoPkg.NewSettlementRepository(),
			ledgerRepoPkg.NewLedgerRepository(),
			groupRepoPkg.NewGroupRepository(),
			groupMemberRepoPkg.NewGroupMemberRepository(),
			groupParticipantRepoPkg.NewGroupParticipantRepository(),
			paymentMethodRepoPkg.NewPaymentMethodRepository(),
			newTestEncryption(t),
		),
		groupID:              uuid.NewString(),
		requesterID:          uuid.NewString(),
		requesterParticipant: uuid.NewString(),
		recipientID:          uuid.NewString(),
		recipientParticipant: uuid.NewString(),
	}

	ctx := context.Background()
	for _, userID := range []string{f.requesterID, f.recipientID} {
		mustExec(t, ctx, `
			INSERT INTO users (id, email, username, full_name, password_hash)
			VALUES ($1, $2, $3, 'Method Test User', 'not-a-real-password')
		`, userID, userID+"@example.test", "pm_"+userID[:8])
	}
	mustExec(t, ctx, `INSERT INTO groups (id, name, created_by) VALUES ($1, 'Method Group', $2)`, f.groupID, f.requesterID)
	mustExec(t, ctx, `INSERT INTO group_members (group_id, user_id, role, status) VALUES ($1, $2, 'owner', 'active')`, f.groupID, f.requesterID)
	mustExec(t, ctx, `INSERT INTO group_members (group_id, user_id, role, status) VALUES ($1, $2, 'member', 'active')`, f.groupID, f.recipientID)
	mustExec(t, ctx, `
		INSERT INTO group_participants (id, group_id, user_id, display_name, participant_type, created_by)
		VALUES ($1, $2, $3, 'Requester', 'registered', $3)
	`, f.requesterParticipant, f.groupID, f.requesterID)
	mustExec(t, ctx, `
		INSERT INTO group_participants (id, group_id, user_id, display_name, participant_type, created_by)
		VALUES ($1, $2, $3, 'Recipient', 'registered', $4)
	`, f.recipientParticipant, f.groupID, f.recipientID, f.requesterID)

	t.Cleanup(func() { f.cleanup(t) })
	return f
}

func (f *recipientMethodFixture) cleanup(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	userIDs := append([]string{f.requesterID, f.recipientID}, f.extraUsers...)
	statements := []struct {
		query string
		args  []any
	}{
		{`DELETE FROM account_ledger WHERE group_id = $1`, []any{f.groupID}},
		{`DELETE FROM settlements WHERE group_id = $1`, []any{f.groupID}},
		{`DELETE FROM group_participants WHERE group_id = $1`, []any{f.groupID}},
		{`DELETE FROM group_members WHERE group_id = $1`, []any{f.groupID}},
		{`DELETE FROM groups WHERE id = $1`, []any{f.groupID}},
		{`DELETE FROM payment_methods WHERE user_id = ANY($1)`, []any{userIDs}},
		{`DELETE FROM users WHERE id = ANY($1)`, []any{userIDs}},
	}
	for _, s := range statements {
		if _, err := integrationPool.Exec(ctx, s.query, s.args...); err != nil {
			t.Errorf("cleanup failed: %v", err)
		}
	}
}

func (f *recipientMethodFixture) addMethod(t *testing.T, userID, visibility, status string) string {
	t.Helper()
	id := uuid.NewString()
	mustExec(t, context.Background(), `
		INSERT INTO payment_methods (id, user_id, method_type, provider_name, visibility, status)
		VALUES ($1, $2, 'bank', $3, $4, $5)
	`, id, userID, "provider_"+id[:6], visibility, status)
	return id
}

func (f *recipientMethodFixture) makeRequesterOwe(t *testing.T) {
	mustExec(t, context.Background(), `
		INSERT INTO account_ledger (group_id, from_participant_id, to_participant_id, source_type, source_id, amount)
		VALUES ($1, $2, $3, $4, $5, 100)
	`, f.groupID, f.requesterParticipant, f.recipientParticipant, ledgerConstants.SourceTypeAdjustment, uuid.NewString())
}

func (f *recipientMethodFixture) list(t *testing.T, requesterID, recipientParticipant string) map[string]bool {
	t.Helper()
	methods, err := f.service.GetRecipientPaymentMethods(context.Background(), requesterID, f.groupID, recipientParticipant)
	if err != nil {
		t.Fatalf("GetRecipientPaymentMethods() error = %v", err)
	}
	seen := make(map[string]bool)
	for _, m := range methods {
		seen[m.ID] = true
	}
	return seen
}

func TestRecipientPaymentMethodsVisibility(t *testing.T) {
	f := newRecipientMethodFixture(t)

	groupMemberID := f.addMethod(t, f.recipientID, "group_members", "active")
	privateID := f.addMethod(t, f.recipientID, "private", "active")
	debtorID := f.addMethod(t, f.recipientID, "debtor_only", "active")
	hiddenID := f.addMethod(t, f.recipientID, "group_members", "hidden")

	before := f.list(t, f.requesterID, f.recipientParticipant)
	if !before[groupMemberID] {
		t.Fatal("group_members method should be visible to member")
	}
	if before[privateID] {
		t.Fatal("private method must not be visible")
	}
	if before[debtorID] {
		t.Fatal("debtor_only method must not be visible before owing")
	}
	if before[hiddenID] {
		t.Fatal("hidden method must not be visible")
	}

	f.makeRequesterOwe(t)

	after := f.list(t, f.requesterID, f.recipientParticipant)
	if !after[groupMemberID] || !after[debtorID] {
		t.Fatalf("after owing, want group_members + debtor_only visible; got %v", after)
	}
	if after[privateID] {
		t.Fatal("private still must not be visible")
	}
}

func TestRecipientPaymentMethodsGuestEmpty(t *testing.T) {
	f := newRecipientMethodFixture(t)
	guest := uuid.NewString()
	mustExec(t, context.Background(), `
		INSERT INTO group_participants (id, group_id, user_id, display_name, participant_type, created_by)
		VALUES ($1, $2, NULL, 'Guest', 'guest', $3)
	`, guest, f.groupID, f.requesterID)

	if got := f.list(t, f.requesterID, guest); len(got) != 0 {
		t.Fatalf("guest recipient should have no methods, got %v", got)
	}
}

func TestRecipientPaymentMethodsRejects(t *testing.T) {
	t.Run("non-member requester", func(t *testing.T) {
		f := newRecipientMethodFixture(t)
		outsider := uuid.NewString()
		mustExec(t, context.Background(), `
			INSERT INTO users (id, email, username, full_name, password_hash)
			VALUES ($1, $2, $3, 'Outsider', 'x')
		`, outsider, outsider+"@example.test", "out_"+outsider[:8])
		f.extraUsers = append(f.extraUsers, outsider)

		_, err := f.service.GetRecipientPaymentMethods(context.Background(), outsider, f.groupID, f.recipientParticipant)
		if !errors.Is(err, expenseConstants.ErrForbiddenGroupAccess) {
			t.Fatalf("error = %v, want %v", err, expenseConstants.ErrForbiddenGroupAccess)
		}
	})

	t.Run("recipient not in group", func(t *testing.T) {
		f := newRecipientMethodFixture(t)
		_, err := f.service.GetRecipientPaymentMethods(context.Background(), f.requesterID, f.groupID, uuid.NewString())
		if !errors.Is(err, settlementConstants.ErrInvalidSettlementParticipants) {
			t.Fatalf("error = %v, want %v", err, settlementConstants.ErrInvalidSettlementParticipants)
		}
	})
}
