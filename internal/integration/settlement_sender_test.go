package integration_test

import (
	"context"
	"errors"
	"testing"
	"time"

	groupRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/group/repository"
	groupMemberRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/group_member/repository"
	groupParticipantRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/group_participant/repository"
	ledgerConstants "github.com/KejarBahasa/kejarbill-api/internal/module/ledger/constants"
	ledgerRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/ledger/repository"
	paymentMethodRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/payment_method/repository"
	settlementConstants "github.com/KejarBahasa/kejarbill-api/internal/module/settlement/constants"
	settlementDto "github.com/KejarBahasa/kejarbill-api/internal/module/settlement/dto"
	settlementRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/settlement/repository"
	settlementServicePkg "github.com/KejarBahasa/kejarbill-api/internal/module/settlement/service"
	"github.com/google/uuid"
)

type senderFixture struct {
	service *settlementServicePkg.SettlementService

	groupID    string
	ownerUser  string
	memberUser string
	stranger   string

	ownerP    string
	memberP   string
	strangerP string
	guestP    string

	users []string
}

func newSenderFixture(t *testing.T) *senderFixture {
	t.Helper()

	f := &senderFixture{
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
		groupID:    uuid.NewString(),
		ownerUser:  uuid.NewString(),
		memberUser: uuid.NewString(),
		stranger:   uuid.NewString(),
		ownerP:     uuid.NewString(),
		memberP:    uuid.NewString(),
		strangerP:  uuid.NewString(),
		guestP:     uuid.NewString(),
	}
	f.users = []string{f.ownerUser, f.memberUser, f.stranger}

	ctx := context.Background()
	for _, u := range f.users {
		mustExec(t, ctx, `
			INSERT INTO users (id, email, username, full_name, password_hash)
			VALUES ($1, $2, $3, 'Sender Test', 'x')
		`, u, u+"@example.test", "snd_"+u[:8])
	}
	mustExec(t, ctx, `INSERT INTO groups (id, name, created_by) VALUES ($1, 'Sender Group', $2)`, f.groupID, f.ownerUser)
	mustExec(t, ctx, `INSERT INTO group_members (group_id, user_id, role, status) VALUES ($1, $2, 'owner', 'active')`, f.groupID, f.ownerUser)
	mustExec(t, ctx, `INSERT INTO group_members (group_id, user_id, role, status) VALUES ($1, $2, 'member', 'active')`, f.groupID, f.memberUser)
	mustExec(t, ctx, `INSERT INTO group_members (group_id, user_id, role, status) VALUES ($1, $2, 'member', 'active')`, f.groupID, f.stranger)

	participants := []struct {
		id, user string
		typ      string
	}{
		{f.ownerP, f.ownerUser, "registered"},
		{f.memberP, f.memberUser, "registered"},
		{f.strangerP, f.stranger, "registered"},
		{f.guestP, "", "guest"},
	}
	for _, p := range participants {
		if p.user == "" {
			mustExec(t, ctx, `
				INSERT INTO group_participants (id, group_id, user_id, display_name, participant_type, created_by)
				VALUES ($1, $2, NULL, 'Guest', 'guest', $3)
			`, p.id, f.groupID, f.ownerUser)
			continue
		}
		mustExec(t, ctx, `
			INSERT INTO group_participants (id, group_id, user_id, display_name, participant_type, created_by)
			VALUES ($1, $2, $3, 'P', 'registered', $4)
		`, p.id, f.groupID, p.user, f.ownerUser)
	}

	t.Cleanup(func() { f.cleanup(t) })
	return f
}

func (f *senderFixture) cleanup(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	for _, q := range []struct {
		sql  string
		args []any
	}{
		{`DELETE FROM account_ledger WHERE group_id = $1`, []any{f.groupID}},
		{`DELETE FROM settlements WHERE group_id = $1`, []any{f.groupID}},
		{`DELETE FROM idempotency_keys WHERE user_id = ANY($1)`, []any{f.users}},
		{`DELETE FROM group_participants WHERE group_id = $1`, []any{f.groupID}},
		{`DELETE FROM group_members WHERE group_id = $1`, []any{f.groupID}},
		{`DELETE FROM groups WHERE id = $1`, []any{f.groupID}},
		{`DELETE FROM users WHERE id = ANY($1)`, []any{f.users}},
	} {
		if _, err := integrationPool.Exec(ctx, q.sql, q.args...); err != nil {
			t.Errorf("cleanup: %v", err)
		}
	}
}

func (f *senderFixture) owe(t *testing.T, fromP, toP string, amount int64) {
	t.Helper()
	mustExec(t, context.Background(), `
		INSERT INTO account_ledger (group_id, from_participant_id, to_participant_id, source_type, source_id, amount)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, f.groupID, fromP, toP, ledgerConstants.SourceTypeExpense, uuid.NewString(), amount)
}

func (f *senderFixture) create(key, requester, fromP, toP string, amount int64) error {
	body := &settlementDto.CreateSettlementBody{
		FromParticipantID: fromP,
		ToParticipantID:   toP,
		PaymentChannel:    settlementConstants.PaymentChannelCash,
		PaidAt:            time.Now().UTC().Format(time.RFC3339),
		Amount:            amount,
	}
	_, err := f.service.Create(context.Background(), requester, f.groupID, key, body)
	return err
}

func TestSettlementSenderRules(t *testing.T) {
	t.Run("member settles for self", func(t *testing.T) {
		f := newSenderFixture(t)
		f.owe(t, f.memberP, f.ownerP, 100_000)
		if err := f.create("self-"+uuid.NewString(), f.memberUser, f.memberP, f.ownerP, 40_000); err != nil {
			t.Fatalf("self settlement error = %v", err)
		}
	})

	t.Run("member uses another member participant", func(t *testing.T) {
		f := newSenderFixture(t)
		err := f.create("other-"+uuid.NewString(), f.memberUser, f.strangerP, f.ownerP, 40_000)
		if !errors.Is(err, settlementConstants.ErrSettlementSenderNotAllowed) {
			t.Fatalf("error = %v, want %v", err, settlementConstants.ErrSettlementSenderNotAllowed)
		}
	})

	t.Run("owner settles for guest", func(t *testing.T) {
		f := newSenderFixture(t)
		f.owe(t, f.guestP, f.ownerP, 100_000)
		if err := f.create("guest-owner-"+uuid.NewString(), f.ownerUser, f.guestP, f.ownerP, 40_000); err != nil {
			t.Fatalf("owner-settles-guest error = %v", err)
		}
	})

	t.Run("member settles for guest", func(t *testing.T) {
		f := newSenderFixture(t)
		err := f.create("guest-member-"+uuid.NewString(), f.memberUser, f.guestP, f.ownerP, 40_000)
		if !errors.Is(err, settlementConstants.ErrSettlementSenderNotAllowed) {
			t.Fatalf("error = %v, want %v", err, settlementConstants.ErrSettlementSenderNotAllowed)
		}
	})
}
