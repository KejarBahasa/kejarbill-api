package service

import (
	"context"

	expenseConstants "github.com/KejarBahasa/kejarbill-api/internal/module/expense/constants"

	groupRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/group/repository"

	groupMemberRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/group_member/repository"

	participantConstants "github.com/KejarBahasa/kejarbill-api/internal/module/group_participant/constants"
	participantDto "github.com/KejarBahasa/kejarbill-api/internal/module/group_participant/dto"
	participantEntity "github.com/KejarBahasa/kejarbill-api/internal/module/group_participant/entity"
	participantRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/group_participant/repository"

	"github.com/KejarBahasa/kejarbill-api/internal/shared/database"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/utils"

	"github.com/jackc/pgx/v5/pgxpool"
)

type GroupParticipantService struct {
	db *pgxpool.Pool

	groupRepo *groupRepoPkg.GroupRepository

	groupMemberRepo *groupMemberRepoPkg.GroupMemberRepository

	participantRepo *participantRepoPkg.GroupParticipantRepository
}

func NewGroupParticipantService(
	db *pgxpool.Pool,

	groupRepo *groupRepoPkg.GroupRepository,

	groupMemberRepo *groupMemberRepoPkg.GroupMemberRepository,

	participantRepo *participantRepoPkg.GroupParticipantRepository,
) *GroupParticipantService {
	return &GroupParticipantService{
		db:              db,
		groupRepo:       groupRepo,
		groupMemberRepo: groupMemberRepo,
		participantRepo: participantRepo,
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

func (s *GroupParticipantService) GetByGroupID(ctx context.Context, requesterUserID string, groupID string) ([]participantEntity.GroupParticipant, error) {
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

	return participants, nil
}
