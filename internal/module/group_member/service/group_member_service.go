package service

import (
	"context"

	"github.com/KejarBahasa/kejarbill-api/internal/shared/database"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/utils"

	expenseConstants "github.com/KejarBahasa/kejarbill-api/internal/module/expense/constants"

	groupRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/group/repository"

	groupMemberConstants "github.com/KejarBahasa/kejarbill-api/internal/module/group_member/constants"
	groupMemberDto "github.com/KejarBahasa/kejarbill-api/internal/module/group_member/dto"
	groupMemberEntity "github.com/KejarBahasa/kejarbill-api/internal/module/group_member/entity"
	groupMemberRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/group_member/repository"

	groupParticipantConstants "github.com/KejarBahasa/kejarbill-api/internal/module/group_participant/constants"
	groupParticipantEntity "github.com/KejarBahasa/kejarbill-api/internal/module/group_participant/entity"
	groupParticipantRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/group_participant/repository"

	userRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/user/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GroupMemberService struct {
	db *pgxpool.Pool

	groupRepo *groupRepoPkg.GroupRepository

	groupMemberRepo *groupMemberRepoPkg.GroupMemberRepository

	groupParticipantRepo *groupParticipantRepoPkg.GroupParticipantRepository

	userRepo *userRepoPkg.UserRepository
}

func NewGroupMemberService(
	db *pgxpool.Pool,

	groupRepo *groupRepoPkg.GroupRepository,

	groupMemberRepo *groupMemberRepoPkg.GroupMemberRepository,

	groupParticipantRepo *groupParticipantRepoPkg.GroupParticipantRepository,

	userRepo *userRepoPkg.UserRepository,
) *GroupMemberService {

	return &GroupMemberService{
		db: db,

		groupRepo: groupRepo,

		groupMemberRepo: groupMemberRepo,

		groupParticipantRepo: groupParticipantRepo,

		userRepo: userRepo,
	}
}

func (s *GroupMemberService) AddMember(ctx context.Context, requesterUserID string, groupID string, req *groupMemberDto.AddGroupMemberBody) error {
	groupExists, err := s.groupRepo.ExistsByID(ctx, s.db, groupID)
	if err != nil {
		return err
	}

	if !groupExists {
		return expenseConstants.ErrGroupNotFound
	}

	hasAccess, err := s.groupMemberRepo.ExistsActiveMember(ctx, s.db, groupID, requesterUserID)
	if err != nil {
		return err
	}

	if !hasAccess {
		return expenseConstants.ErrForbiddenGroupAccess
	}

	targetUserExists, err := s.userRepo.ExistsByID(ctx, s.db, req.UserID)
	if err != nil {
		return err
	}

	if !targetUserExists {
		return expenseConstants.ErrUserNotFound
	}

	alreadyMember, err := s.groupMemberRepo.ExistsActiveMember(ctx, s.db, groupID, req.UserID)
	if err != nil {
		return err
	}

	if alreadyMember {
		return groupMemberConstants.ErrAlreadyMember
	}

	targetUser, err := s.userRepo.FindByID(ctx, req.UserID)
	if err != nil {
		return err
	}

	return database.WithTransaction(ctx, s.db, func(tx pgx.Tx) error {
		err := s.groupMemberRepo.Create(ctx, tx, &groupMemberEntity.GroupMember{
			GroupID: groupID,
			UserID:  req.UserID,
			Role:    groupMemberConstants.RoleMember,
			Status:  groupMemberConstants.StatusActive,
		})
		if err != nil {
			return err
		}

		err = s.groupParticipantRepo.Create(ctx, tx, &groupParticipantEntity.GroupParticipant{
			GroupID:         groupID,
			UserID:          utils.PtrOrNil(req.UserID),
			ParticipantType: groupParticipantConstants.TypeRegistered,
			DisplayName:     targetUser.Name,
			CreatedBy:       requesterUserID,
		})
		if err != nil {
			return err
		}

		return nil
	})
}

func (s *GroupMemberService) AddMembersBulk(ctx context.Context, requesterUserID string, groupID string, req *groupMemberDto.AddGroupMembersBulkBody) error {
	groupExists, err := s.groupRepo.ExistsByID(ctx, s.db, groupID)
	if err != nil {
		return err
	}
	if !groupExists {
		return expenseConstants.ErrGroupNotFound
	}

	hasAccess, err := s.groupMemberRepo.ExistsActiveMember(ctx, s.db, groupID, requesterUserID)
	if err != nil {
		return err
	}
	if !hasAccess {
		return expenseConstants.ErrForbiddenGroupAccess
	}

	userIDs := utils.UniqueStrings(req.UserIDs)
	users, err := s.userRepo.FindByIDs(ctx, s.db, userIDs)
	if err != nil {
		return err
	}

	/*
		Validate all users exist
	*/
	if len(users) != len(userIDs) {
		return groupMemberConstants.ErrUserNotFound
	}

	existingMembers, err := s.groupMemberRepo.FindActiveMemberUserIDs(ctx, s.db, groupID, userIDs)
	if err != nil {
		return err
	}

	if len(existingMembers) > 0 {
		return groupMemberConstants.ErrSomeUsersAlreadyMember
	}

	memberEntities := make([]groupMemberEntity.GroupMember, 0, len(users))
	participantEntities := make([]groupParticipantEntity.GroupParticipant, 0, len(users))

	for _, user := range users {
		memberEntities = append(memberEntities, groupMemberEntity.GroupMember{
			GroupID: groupID,
			UserID:  user.ID,
			Role:    groupMemberConstants.RoleMember,
			Status:  groupMemberConstants.StatusActive,
		})

		userID := user.ID

		participantEntities = append(
			participantEntities,
			groupParticipantEntity.GroupParticipant{
				GroupID:         groupID,
				UserID:          &userID,
				ParticipantType: groupParticipantConstants.TypeRegistered,
				DisplayName:     user.Name,
				CreatedBy:       requesterUserID,
			},
		)
	}

	return database.WithTransaction(ctx, s.db, func(tx pgx.Tx) error {
		err := s.groupMemberRepo.BulkCreate(ctx, tx, memberEntities)
		if err != nil {
			return err
		}

		err = s.groupParticipantRepo.BulkCreate(ctx, tx, participantEntities)
		if err != nil {
			return err
		}

		return nil
	},
	)
}
