package integration_test

import (
	"context"
	"errors"
	"testing"

	expenseConstants "github.com/KejarBahasa/kejarbill-api/internal/module/expense/constants"
	groupRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/group/repository"
	groupMemberConstants "github.com/KejarBahasa/kejarbill-api/internal/module/group_member/constants"
	groupMemberDto "github.com/KejarBahasa/kejarbill-api/internal/module/group_member/dto"
	groupMemberRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/group_member/repository"
	groupMemberServicePkg "github.com/KejarBahasa/kejarbill-api/internal/module/group_member/service"
	participantConstants "github.com/KejarBahasa/kejarbill-api/internal/module/group_participant/constants"
	participantRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/group_participant/repository"
	participantServicePkg "github.com/KejarBahasa/kejarbill-api/internal/module/group_participant/service"
	userRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/user/repository"
	"github.com/google/uuid"
)

type claimFixture struct {
	service *participantServicePkg.GroupParticipantService

	groupID string
	ownerID string
	guestID string
	userID  string

	extraUsers []string
}

func newClaimFixture(t *testing.T) *claimFixture {
	t.Helper()

	f := &claimFixture{
		service: participantServicePkg.NewGroupParticipantService(
			integrationPool,
			groupRepoPkg.NewGroupRepository(),
			groupMemberRepoPkg.NewGroupMemberRepository(),
			participantRepoPkg.NewGroupParticipantRepository(),
			userRepoPkg.NewUserRepository(integrationPool),
		),
		groupID: uuid.NewString(),
		ownerID: uuid.NewString(),
		guestID: uuid.NewString(),
		userID:  uuid.NewString(),
	}

	ctx := context.Background()

	for _, userID := range []string{f.ownerID, f.userID} {
		mustExec(t, ctx, `
			INSERT INTO users (id, email, username, full_name, password_hash)
			VALUES ($1, $2, $3, 'Claim Test User', 'not-a-real-password')
		`, userID, userID+"@example.test", "claim_"+userID[:8])
	}

	mustExec(t, ctx, `
		INSERT INTO groups (id, name, created_by)
		VALUES ($1, 'Claim Test Group', $2)
	`, f.groupID, f.ownerID)
	mustExec(t, ctx, `
		INSERT INTO group_members (group_id, user_id, role, status)
		VALUES ($1, $2, 'owner', 'active')
	`, f.groupID, f.ownerID)
	mustExec(t, ctx, `
		INSERT INTO group_participants (id, group_id, user_id, display_name, participant_type, created_by)
		VALUES ($1, $2, $3, 'Owner', 'registered', $3)
	`, uuid.NewString(), f.groupID, f.ownerID)
	mustExec(t, ctx, `
		INSERT INTO group_participants (id, group_id, user_id, display_name, participant_type, created_by)
		VALUES ($1, $2, NULL, 'Guest', 'guest', $3)
	`, f.guestID, f.groupID, f.ownerID)

	t.Cleanup(func() { f.cleanup(t) })
	return f
}

func (f *claimFixture) cleanup(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	userIDs := append([]string{f.ownerID, f.userID}, f.extraUsers...)
	statements := []struct {
		query string
		args  []any
	}{
		{`DELETE FROM group_participants WHERE group_id = $1`, []any{f.groupID}},
		{`DELETE FROM group_members WHERE group_id = $1`, []any{f.groupID}},
		{`DELETE FROM groups WHERE id = $1`, []any{f.groupID}},
		{`DELETE FROM users WHERE id = ANY($1)`, []any{userIDs}},
	}
	for _, s := range statements {
		if _, err := integrationPool.Exec(ctx, s.query, s.args...); err != nil {
			t.Errorf("cleanup failed: %v", err)
		}
	}
}

// seedPlainMember adds a new registered user with role 'member' in the group and returns its user id.
func (f *claimFixture) seedPlainMember(t *testing.T) string {
	t.Helper()
	userID := uuid.NewString()
	ctx := context.Background()
	mustExec(t, ctx, `
		INSERT INTO users (id, email, username, full_name, password_hash)
		VALUES ($1, $2, $3, 'Plain Member', 'not-a-real-password')
	`, userID, userID+"@example.test", "member_"+userID[:8])
	mustExec(t, ctx, `
		INSERT INTO group_members (group_id, user_id, role, status)
		VALUES ($1, $2, 'member', 'active')
	`, f.groupID, userID)
	f.extraUsers = append(f.extraUsers, userID)
	return userID
}

func (f *claimFixture) makeTargetMember(t *testing.T) {
	t.Helper()
	mustExec(t, context.Background(), `
		INSERT INTO group_members (group_id, user_id, role, status)
		VALUES ($1, $2, 'member', 'active')
	`, f.groupID, f.userID)
}

func (f *claimFixture) makeTargetParticipant(t *testing.T) {
	t.Helper()
	mustExec(t, context.Background(), `
		INSERT INTO group_participants (id, group_id, user_id, display_name, participant_type, created_by)
		VALUES ($1, $2, $3, 'Target', 'registered', $4)
	`, uuid.NewString(), f.groupID, f.userID, f.ownerID)
}

func (f *claimFixture) claimedUser(t *testing.T) (string, bool, bool, string) {
	t.Helper()
	var userID string
	var claimed bool
	var isRegistered bool
	var displayName string
	err := integrationPool.QueryRow(context.Background(), `
		SELECT COALESCE(user_id::text, ''), claimed_at IS NOT NULL, participant_type = 'registered', display_name
		FROM group_participants
		WHERE id = $1
	`, f.guestID).Scan(&userID, &claimed, &isRegistered, &displayName)
	if err != nil {
		t.Fatalf("query claimed: %v", err)
	}
	return userID, claimed, isRegistered, displayName
}

func (f *claimFixture) isMember(t *testing.T, userID string) bool {
	t.Helper()
	return countRows(t, context.Background(), `
		SELECT COUNT(*) FROM group_members
		WHERE group_id = $1 AND user_id = $2 AND status = 'active' AND role = 'member'
	`, f.groupID, userID) == 1
}

func TestGroupParticipantClaimGuest(t *testing.T) {
	fixture := newClaimFixture(t)

	err := fixture.service.ClaimGuestParticipant(context.Background(), fixture.ownerID, fixture.groupID, fixture.guestID, fixture.userID)
	if err != nil {
		t.Fatalf("ClaimGuestParticipant() error = %v", err)
	}

	userID, claimed, isRegistered, displayName := fixture.claimedUser(t)
	if userID != fixture.userID {
		t.Fatalf("linked user_id = %q, want %q", userID, fixture.userID)
	}
	if !claimed || !isRegistered {
		t.Fatalf("participant not flipped to claimed registered: claimed=%t registered=%t", claimed, isRegistered)
	}
	if displayName != "Claim Test User" {
		t.Fatalf("display_name = %q, want %q", displayName, "Claim Test User")
	}
	if !fixture.isMember(t, fixture.userID) {
		t.Fatal("target user was not added as active member")
	}
}

func TestGroupParticipantClaimGuestRequiresManagerRole(t *testing.T) {
	fixture := newClaimFixture(t)
	plainMemberID := fixture.seedPlainMember(t)

	err := fixture.service.ClaimGuestParticipant(context.Background(), plainMemberID, fixture.groupID, fixture.guestID, fixture.userID)
	if !errors.Is(err, groupMemberConstants.ErrForbiddenGroupRole) {
		t.Fatalf("ClaimGuestParticipant() error = %v, want %v", err, groupMemberConstants.ErrForbiddenGroupRole)
	}

	_, claimed, _, _ := fixture.claimedUser(t)
	if claimed {
		t.Fatal("guest was claimed despite insufficient role")
	}
}

func TestGroupParticipantClaimGuestRejects(t *testing.T) {
	tests := []struct {
		name    string
		arrange func(t *testing.T, f *claimFixture) (participantID, targetUserID string)
		wantErr error
	}{
		{
			name: "already member",
			arrange: func(t *testing.T, f *claimFixture) (string, string) {
				f.makeTargetMember(t)
				return f.guestID, f.userID
			},
			wantErr: groupMemberConstants.ErrAlreadyMember,
		},
		{
			name: "already participant",
			arrange: func(t *testing.T, f *claimFixture) (string, string) {
				f.makeTargetParticipant(t)
				return f.guestID, f.userID
			},
			wantErr: groupMemberConstants.ErrAlreadyMember,
		},
		{
			name: "target not found",
			arrange: func(_ *testing.T, f *claimFixture) (string, string) {
				return f.guestID, uuid.NewString()
			},
			wantErr: expenseConstants.ErrUserNotFound,
		},
		{
			name: "participant not found",
			arrange: func(_ *testing.T, f *claimFixture) (string, string) {
				return uuid.NewString(), f.userID
			},
			wantErr: participantConstants.ErrParticipantNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := newClaimFixture(t)
			participantID, targetUserID := test.arrange(t, fixture)

			err := fixture.service.ClaimGuestParticipant(context.Background(), fixture.ownerID, fixture.groupID, participantID, targetUserID)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("ClaimGuestParticipant() error = %v, want %v", err, test.wantErr)
			}
		})
	}
}

func TestGroupParticipantClaimGuestDoubleClaim(t *testing.T) {
	fixture := newClaimFixture(t)

	if err := fixture.service.ClaimGuestParticipant(context.Background(), fixture.ownerID, fixture.groupID, fixture.guestID, fixture.userID); err != nil {
		t.Fatalf("first claim error = %v", err)
	}

	err := fixture.service.ClaimGuestParticipant(context.Background(), fixture.ownerID, fixture.groupID, fixture.guestID, uuid.NewString())
	if !errors.Is(err, participantConstants.ErrParticipantNotClaimable) {
		t.Fatalf("second claim error = %v, want %v", err, participantConstants.ErrParticipantNotClaimable)
	}
}

func TestGroupMemberAddRequiresManagerRole(t *testing.T) {
	fixture := newClaimFixture(t)
	svc := groupMemberServicePkg.NewGroupMemberService(
		integrationPool,
		groupRepoPkg.NewGroupRepository(),
		groupMemberRepoPkg.NewGroupMemberRepository(),
		participantRepoPkg.NewGroupParticipantRepository(),
		userRepoPkg.NewUserRepository(integrationPool),
	)
	body := &groupMemberDto.AddGroupMemberBody{UserID: fixture.userID}

	plainMemberID := fixture.seedPlainMember(t)
	if err := svc.AddMember(context.Background(), plainMemberID, fixture.groupID, body); !errors.Is(err, groupMemberConstants.ErrForbiddenGroupRole) {
		t.Fatalf("AddMember() as member error = %v, want %v", err, groupMemberConstants.ErrForbiddenGroupRole)
	}

	if err := svc.AddMember(context.Background(), fixture.ownerID, fixture.groupID, body); err != nil {
		t.Fatalf("AddMember() as owner error = %v", err)
	}
}
