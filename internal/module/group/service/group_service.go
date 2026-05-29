package service

import (
	"context"

	expenseConstants "github.com/KejarBahasa/kejarbill-api/internal/module/expense/constants"

	groupEntity "github.com/KejarBahasa/kejarbill-api/internal/module/group/entity"
	groupRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/group/repository"

	groupMemberConstants "github.com/KejarBahasa/kejarbill-api/internal/module/group_member/constants"
	groupMemberEntity "github.com/KejarBahasa/kejarbill-api/internal/module/group_member/entity"
	groupMemberRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/group_member/repository"

	groupParticipantConstants "github.com/KejarBahasa/kejarbill-api/internal/module/group_participant/constants"
	groupParticipantEntity "github.com/KejarBahasa/kejarbill-api/internal/module/group_participant/entity"
	groupParticipantRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/group_participant/repository"

	userRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/user/repository"

	"github.com/KejarBahasa/kejarbill-api/internal/shared/database"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/utils"

	"github.com/jackc/pgx/v5/pgxpool"
)

type GroupService struct {
	db *pgxpool.Pool

	groupRepo *groupRepoPkg.GroupRepository

	groupMemberRepo *groupMemberRepoPkg.GroupMemberRepository

	groupParticipantRepo *groupParticipantRepoPkg.GroupParticipantRepository

	userRepo *userRepoPkg.UserRepository
}

func NewGroupService(
	db *pgxpool.Pool,

	groupRepo *groupRepoPkg.GroupRepository,

	groupMemberRepo *groupMemberRepoPkg.GroupMemberRepository,

	groupParticipantRepo *groupParticipantRepoPkg.GroupParticipantRepository,

	userRepo *userRepoPkg.UserRepository,
) *GroupService {

	return &GroupService{
		db: db,

		groupRepo: groupRepo,

		groupMemberRepo: groupMemberRepo,

		groupParticipantRepo: groupParticipantRepo,

		userRepo: userRepo,
	}
}

func (s *GroupService) Create(ctx context.Context, userID string, name string, description *string) (string, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return "", err
	}

	var groupID string
	err = database.WithTransaction(ctx, s.db, func(tx database.PgxExt) error {
		createdGroupID, err := s.groupRepo.Create(ctx, tx, &groupEntity.Group{
			Name:        name,
			Description: description,
			CreatedBy:   userID,
		})
		if err != nil {
			return err
		}

		groupID = createdGroupID

		err = s.groupMemberRepo.Create(ctx, tx, &groupMemberEntity.GroupMember{
			GroupID: groupID,
			UserID:  userID,
			Role:    groupMemberConstants.RoleOwner,
			Status:  groupMemberConstants.StatusActive,
		})
		if err != nil {
			return err
		}

		err = s.groupParticipantRepo.Create(ctx, tx, &groupParticipantEntity.GroupParticipant{
			GroupID:         groupID,
			UserID:          utils.PtrOrNil(userID),
			DisplayName:     user.Name,
			ParticipantType: groupParticipantConstants.TypeRegistered,
			CreatedBy:       userID,
		})
		if err != nil {
			return err
		}

		return nil
	})

	return groupID, err
}

func (s *GroupService) GetDetailByID(ctx context.Context, requesterUserID string, groupID string) (*groupEntity.GroupDetail, error) {
	groupExists, err := s.groupRepo.ExistsByID(ctx, s.db, groupID)
	if err != nil {
		return nil, err
	}
	if !groupExists {
		return nil, expenseConstants.ErrGroupNotFound
	}

	hasAccess, err := s.groupMemberRepo.ExistsActiveMember(ctx, s.db, groupID, requesterUserID)
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, expenseConstants.ErrForbiddenGroupAccess
	}

	group, err := s.groupRepo.FindDetailByID(ctx, s.db, groupID)
	if err != nil {
		return nil, err
	}

	return group, nil
}
