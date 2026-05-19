package service

import (
	"context"

	groupEntity "github.com/KejarBahasa/kejarbill-api/internal/module/group/entity"
	groupRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/group/repository"

	participantEntity "github.com/KejarBahasa/kejarbill-api/internal/module/participant/entity"
	participantRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/participant/repository"

	"github.com/KejarBahasa/kejarbill-api/internal/shared/database"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GroupService struct {
	db *pgxpool.Pool

	groupRepo *groupRepoPkg.GroupRepository

	participantRepo *participantRepoPkg.ParticipantRepository
}

func NewGroupService(
	db *pgxpool.Pool,

	groupRepo *groupRepoPkg.GroupRepository,

	participantRepo *participantRepoPkg.ParticipantRepository,
) *GroupService {

	return &GroupService{
		db: db,

		groupRepo: groupRepo,

		participantRepo: participantRepo,
	}
}

func (s *GroupService) Create(ctx context.Context, userID string, name string) (string, error) {
	var groupID string
	err := database.WithTransaction(ctx, s.db, func(tx pgx.Tx) error {
		createdGroupID, err := s.groupRepo.Create(ctx, tx, &groupEntity.Group{
			Name:      name,
			CreatedBy: userID,
		})
		if err != nil {
			return err
		}

		groupID = createdGroupID
		err = s.participantRepo.Create(ctx, tx, &participantEntity.Participant{
			GroupID:         groupID,
			UserID:          userID,
			DisplayName:     "You",
			ParticipantType: "registered",
		})
		if err != nil {
			return err
		}

		return nil
	})

	return groupID, err
}
