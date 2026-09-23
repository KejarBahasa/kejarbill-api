package integration_test

import (
	"context"
	"testing"

	groupRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/group/repository"
	groupMemberRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/group_member/repository"
	participantRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/group_participant/repository"
	participantServicePkg "github.com/KejarBahasa/kejarbill-api/internal/module/group_participant/service"
	userRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/user/repository"
)

func TestGroupParticipantsIncludeUsername(t *testing.T) {
	f := newSummaryFixture(t)
	service := participantServicePkg.NewGroupParticipantService(
		integrationPool,
		groupRepoPkg.NewGroupRepository(),
		groupMemberRepoPkg.NewGroupMemberRepository(),
		participantRepoPkg.NewGroupParticipantRepository(),
		userRepoPkg.NewUserRepository(integrationPool),
	)

	participants, err := service.GetByGroupID(context.Background(), f.memberUser, f.groupID)
	if err != nil {
		t.Fatalf("GetByGroupID() error = %v", err)
	}
	if len(participants) != 3 {
		t.Fatalf("participants length = %d, want 3", len(participants))
	}

	byID := make(map[string]struct {
		username *string
		isSelf   bool
		role     *string
	})
	for _, participant := range participants {
		byID[participant.ID] = struct {
			username *string
			isSelf   bool
			role     *string
		}{participant.Username, participant.IsSelf, participant.Role}
	}

	member := byID[f.memberP]
	if member.username == nil || *member.username != "sm_"+f.memberUser[:8] || !member.isSelf || member.role == nil || *member.role != "member" {
		t.Fatalf("member response = %+v, want username/self/member role", member)
	}
	guest := byID[f.guestP]
	if guest.username != nil || guest.isSelf || guest.role != nil {
		t.Fatalf("guest response = %+v, want null username/role and false self", guest)
	}
}
