package integration_test

import (
	"context"
	"errors"
	"testing"

	expenseConstants "github.com/KejarBahasa/kejarbill-api/internal/module/expense/constants"
	expenseRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/expense/repository"
	groupRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/group/repository"
	groupServicePkg "github.com/KejarBahasa/kejarbill-api/internal/module/group/service"
	groupMemberRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/group_member/repository"
	groupParticipantRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/group_participant/repository"
	ledgerConstants "github.com/KejarBahasa/kejarbill-api/internal/module/ledger/constants"
	ledgerRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/ledger/repository"
	userRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/user/repository"
	"github.com/google/uuid"
)

type summaryFixture struct {
	service *groupServicePkg.GroupService

	groupID     string
	ownerUser   string
	memberUser  string
	ownerP      string
	memberP     string
	guestP      string
	outsider    string
	outsiderExp string

	users []string
}

func newSummaryFixture(t *testing.T) *summaryFixture {
	t.Helper()

	f := &summaryFixture{
		service: groupServicePkg.NewGroupService(
			integrationPool,
			groupRepoPkg.NewGroupRepository(),
			groupMemberRepoPkg.NewGroupMemberRepository(),
			groupParticipantRepoPkg.NewGroupParticipantRepository(),
			userRepoPkg.NewUserRepository(integrationPool),
			expenseRepoPkg.NewExpenseRepository(),
			ledgerRepoPkg.NewLedgerRepository(),
		),
		groupID:     uuid.NewString(),
		ownerUser:   uuid.NewString(),
		memberUser:  uuid.NewString(),
		ownerP:      uuid.NewString(),
		memberP:     uuid.NewString(),
		guestP:      uuid.NewString(),
		outsider:    uuid.NewString(),
		outsiderExp: uuid.NewString(),
	}
	f.users = []string{f.ownerUser, f.memberUser, f.outsider, f.outsiderExp}

	ctx := context.Background()
	for _, u := range f.users {
		mustExec(t, ctx, `
			INSERT INTO users (id, email, username, full_name, password_hash)
			VALUES ($1, $2, $3, 'Summary Test', 'x')
		`, u, u+"@example.test", "sm_"+u[:8])
	}
	mustExec(t, ctx, `INSERT INTO groups (id, name, created_by) VALUES ($1, 'Summary Group', $2)`, f.groupID, f.ownerUser)
	mustExec(t, ctx, `INSERT INTO group_members (group_id, user_id, role, status) VALUES ($1, $2, 'owner', 'active')`, f.groupID, f.ownerUser)
	mustExec(t, ctx, `INSERT INTO group_members (group_id, user_id, role, status) VALUES ($1, $2, 'member', 'active')`, f.groupID, f.memberUser)
	mustExec(t, ctx, `
		INSERT INTO group_participants (id, group_id, user_id, display_name, participant_type, created_by)
		VALUES ($1, $2, $3, 'Owner', 'registered', $3)
	`, f.ownerP, f.groupID, f.ownerUser)
	mustExec(t, ctx, `
		INSERT INTO group_participants (id, group_id, user_id, display_name, participant_type, created_by)
		VALUES ($1, $2, $3, 'Member', 'registered', $4)
	`, f.memberP, f.groupID, f.memberUser, f.ownerUser)
	mustExec(t, ctx, `
		INSERT INTO group_participants (id, group_id, user_id, display_name, participant_type, created_by)
		VALUES ($1, $2, NULL, 'Guest', 'guest', $3)
	`, f.guestP, f.groupID, f.ownerUser)

	t.Cleanup(func() { f.cleanup(t) })
	return f
}

func (f *summaryFixture) cleanup(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	for _, q := range []struct {
		sql  string
		args []any
	}{
		{`DELETE FROM settlements WHERE group_id = $1`, []any{f.groupID}},
		{`DELETE FROM account_ledger WHERE group_id = $1`, []any{f.groupID}},
		{`DELETE FROM expense_participants WHERE expense_id IN (SELECT id FROM expenses WHERE group_id = $1)`, []any{f.groupID}},
		{`DELETE FROM expenses WHERE group_id = $1`, []any{f.groupID}},
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

func (f *summaryFixture) addExpense(t *testing.T, payer string, total int64) {
	t.Helper()
	mustExec(t, context.Background(), `
		INSERT INTO expenses (id, group_id, title, paid_by_participant_id, total_amount, split_method, expense_date, created_by)
		VALUES ($1, $2, 'exp', $3, $4, 'custom', NOW(), $5)
	`, uuid.NewString(), f.groupID, payer, total, f.ownerUser)
}

func (f *summaryFixture) addLedger(t *testing.T, from, to string, amount int64) {
	t.Helper()
	mustExec(t, context.Background(), `
		INSERT INTO account_ledger (group_id, from_participant_id, to_participant_id, source_type, source_id, amount)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, f.groupID, from, to, ledgerConstants.SourceTypeAdjustment, uuid.NewString(), amount)
}

func TestGroupSummaryEmpty(t *testing.T) {
	f := newSummaryFixture(t)

	s, err := f.service.GetSummary(context.Background(), f.memberUser, f.groupID)
	if err != nil {
		t.Fatalf("GetSummary() error = %v", err)
	}
	if s.GroupTotal != 0 || s.MyTotalPaid != 0 || s.MyTotalDebt != 0 || s.MyTotalCredit != 0 {
		t.Fatalf("expected all zero, got %+v", s)
	}
}

func TestGroupSummaryAggregates(t *testing.T) {
	f := newSummaryFixture(t)

	// group total = 140000; member paid one expense of 40000.
	f.addExpense(t, f.ownerP, 100_000)
	f.addExpense(t, f.memberP, 40_000)

	// netting: member owes owner 50k, owner owes member 20k => member debt 30k
	f.addLedger(t, f.memberP, f.ownerP, 50_000)
	f.addLedger(t, f.ownerP, f.memberP, 20_000)
	// guest owes member 10k => member credit 10k
	f.addLedger(t, f.guestP, f.memberP, 10_000)

	s, err := f.service.GetSummary(context.Background(), f.memberUser, f.groupID)
	if err != nil {
		t.Fatalf("GetSummary() error = %v", err)
	}
	if s.GroupTotal != 140_000 {
		t.Fatalf("group_total = %d, want 140000", s.GroupTotal)
	}
	if s.MyTotalPaid != 40_000 {
		t.Fatalf("my_total_paid = %d, want 40000", s.MyTotalPaid)
	}
	if s.MyTotalDebt != 30_000 {
		t.Fatalf("my_total_debt = %d, want 30000", s.MyTotalDebt)
	}
	if s.MyTotalCredit != 10_000 {
		t.Fatalf("my_total_credit = %d, want 10000", s.MyTotalCredit)
	}
}

func TestGroupSummaryRejectsNonMember(t *testing.T) {
	f := newSummaryFixture(t)

	_, err := f.service.GetSummary(context.Background(), f.outsider, f.groupID)
	if !errors.Is(err, expenseConstants.ErrForbiddenGroupAccess) {
		t.Fatalf("error = %v, want %v", err, expenseConstants.ErrForbiddenGroupAccess)
	}
}

func TestGroupSummaryMemberWithoutParticipant(t *testing.T) {
	f := newSummaryFixture(t)
	// outsiderExp is an active member but has NO registered participant.
	mustExec(t, context.Background(), `INSERT INTO group_members (group_id, user_id, role, status) VALUES ($1, $2, 'member', 'active')`, f.groupID, f.outsiderExp)
	f.addExpense(t, f.ownerP, 70_000)

	s, err := f.service.GetSummary(context.Background(), f.outsiderExp, f.groupID)
	if err != nil {
		t.Fatalf("GetSummary() error = %v", err)
	}
	if s.GroupTotal != 70_000 {
		t.Fatalf("group_total = %d, want 70000", s.GroupTotal)
	}
	if s.MyTotalPaid != 0 || s.MyTotalDebt != 0 || s.MyTotalCredit != 0 {
		t.Fatalf("expected my_* zero for participant-less member, got %+v", s)
	}
}
