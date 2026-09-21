package service

import (
	"context"
	"errors"

	expenseConstants "github.com/KejarBahasa/kejarbill-api/internal/module/expense/constants"

	groupRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/group/repository"

	groupMemberConstants "github.com/KejarBahasa/kejarbill-api/internal/module/group_member/constants"
	groupMemberEntity "github.com/KejarBahasa/kejarbill-api/internal/module/group_member/entity"
	groupMemberRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/group_member/repository"

	userRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/user/repository"

	participantConstants "github.com/KejarBahasa/kejarbill-api/internal/module/group_participant/constants"
	participantDto "github.com/KejarBahasa/kejarbill-api/internal/module/group_participant/dto"
	participantEntity "github.com/KejarBahasa/kejarbill-api/internal/module/group_participant/entity"
	participantRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/group_participant/repository"

	"github.com/KejarBahasa/kejarbill-api/internal/shared/database"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/utils"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GroupParticipantService struct {
	db *pgxpool.Pool

	groupRepo *groupRepoPkg.GroupRepository

	groupMemberRepo *groupMemberRepoPkg.GroupMemberRepository

	participantRepo *participantRepoPkg.GroupParticipantRepository

	userRepo *userRepoPkg.UserRepository
}

func NewGroupParticipantService(
	db *pgxpool.Pool,

	groupRepo *groupRepoPkg.GroupRepository,

	groupMemberRepo *groupMemberRepoPkg.GroupMemberRepository,

	participantRepo *participantRepoPkg.GroupParticipantRepository,

	userRepo *userRepoPkg.UserRepository,
) *GroupParticipantService {
	return &GroupParticipantService{
		db:              db,
		groupRepo:       groupRepo,
		groupMemberRepo: groupMemberRepo,
		participantRepo: participantRepo,
		userRepo:        userRepo,
	}
}

func (s *GroupParticipantService) CreateGuestParticipants(ctx context.Context, requesterUserID string, groupID string, req *participantDto.CreateGuestParticipantsBody) error {
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

	displayNames := make([]string, 0)

	for _, guest := range req.Guests {
		displayNames = append(displayNames, guest.DisplayName)
	}

	uniqueNames := utils.UniqueStrings(displayNames)

	if len(uniqueNames) != len(displayNames) {
		return participantConstants.ErrDuplicateGuestName
	}

	participants := make([]participantEntity.GroupParticipant, 0, len(req.Guests))

	for _, guest := range req.Guests {
		participants = append(participants, participantEntity.GroupParticipant{
			GroupID:         groupID,
			UserID:          nil,
			ParticipantType: participantConstants.TypeGuest,
			DisplayName:     guest.DisplayName,
			CreatedBy:       requesterUserID,
		})
	}

	return database.WithTransaction(ctx, s.db, func(tx database.PgxExt) error {
		return s.participantRepo.BulkCreate(ctx, tx, participants)
	})
}

func (s *GroupParticipantService) GetByGroupID(ctx context.Context, requesterUserID string, groupID string) ([]participantDto.ParticipantResponse, error) {
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

	participants, err := s.participantRepo.FindByGroupID(ctx, s.db, groupID)
	if err != nil {
		return nil, err
	}

	roles, err := s.groupMemberRepo.FindActiveMemberRolesByGroup(ctx, s.db, groupID)
	if err != nil {
		return nil, err
	}

	result := make([]participantDto.ParticipantResponse, 0, len(participants))
	for _, participant := range participants {
		var role *string
		isSelf := false
		if participant.UserID != nil {
			if r, ok := roles[*participant.UserID]; ok {
				role = &r
			}
			isSelf = *participant.UserID == requesterUserID
		}

		result = append(result, participantDto.ParticipantResponse{
			ID:              participant.ID,
			UserID:          participant.UserID,
			ParticipantType: participant.ParticipantType,
			DisplayName:     participant.DisplayName,
			Role:            role,
			IsSelf:          isSelf,
		})
	}

	return result, nil
}

func (s *GroupParticipantService) ClaimGuestParticipant(ctx context.Context, requesterUserID string, groupID string, participantID string, targetUserID string) error {
	groupExists, err := s.groupRepo.ExistsByID(ctx, s.db, groupID)
	if err != nil {
		return err
	}
	if !groupExists {
		return expenseConstants.ErrGroupNotFound
	}

	requesterRole, err := s.groupMemberRepo.FindActiveMemberRole(ctx, s.db, groupID, requesterUserID)
	if err != nil {
		return err
	}
	if requesterRole == "" {
		return expenseConstants.ErrForbiddenGroupAccess
	}
	if !groupMemberConstants.CanManage(requesterRole) {
		return groupMemberConstants.ErrForbiddenGroupRole
	}

	participant, err := s.participantRepo.FindByIDAndGroupID(ctx, s.db, participantID, groupID)
	if err != nil {
		return err
	}
	if participant == nil {
		return participantConstants.ErrParticipantNotFound
	}
	if participant.ParticipantType != participantConstants.TypeGuest || participant.UserID != nil {
		return participantConstants.ErrParticipantNotClaimable
	}

	targetUser, err := s.userRepo.FindByID(ctx, targetUserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return expenseConstants.ErrUserNotFound
		}
		return err
	}

	displayName := targetUser.Name
	if displayName == "" {
		displayName = participant.DisplayName
	}

	alreadyMember, err := s.groupMemberRepo.ExistsActiveMember(ctx, s.db, groupID, targetUserID)
	if err != nil {
		return err
	}
	if alreadyMember {
		return groupMemberConstants.ErrAlreadyMember
	}

	alreadyParticipant, err := s.participantRepo.ExistsByUserIDAndGroupID(ctx, s.db, targetUserID, groupID)
	if err != nil {
		return err
	}
	if alreadyParticipant {
		return groupMemberConstants.ErrAlreadyMember
	}

	return database.WithTransaction(ctx, s.db, func(tx database.PgxExt) error {
		affected, err := s.participantRepo.ClaimGuest(ctx, tx, participantID, groupID, targetUserID, displayName)
		if err != nil {
			return err
		}
		if affected == 0 {
			return participantConstants.ErrParticipantNotClaimable
		}

		return s.groupMemberRepo.Create(ctx, tx, &groupMemberEntity.GroupMember{
			GroupID: groupID,
			UserID:  targetUserID,
			Role:    groupMemberConstants.RoleMember,
			Status:  groupMemberConstants.StatusActive,
		})
	})
}
