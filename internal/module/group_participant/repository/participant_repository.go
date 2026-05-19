package repository

import (
	"context"

	"github.com/KejarBahasa/kejarbill-api/internal/module/group_participant/entity"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/database"
)

type GroupParticipantRepository struct {
}

func NewGroupParticipantRepository() *GroupParticipantRepository {
	return &GroupParticipantRepository{}
}

func (r *GroupParticipantRepository) CountByIDsAndGroupID(ctx context.Context, db database.PgxExt, groupID string, participantIDs []string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM group_participants
		WHERE
			group_id = $1
			AND id = ANY($2)
	`

	var count int
	err := db.QueryRow(ctx, query, groupID, participantIDs).Scan(&count)

	return count, err
}

func (r *GroupParticipantRepository) ExistsByIDAndGroupID(ctx context.Context, db database.PgxExt, groupID string, participantID string) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM group_participants
			WHERE
				id = $1
				AND group_id = $2
		)
	`

	var exists bool
	err := db.QueryRow(ctx, query, participantID, groupID).Scan(&exists)

	return exists, err
}

func (r *GroupParticipantRepository) ExistsByUserIDAndGroupID(ctx context.Context, db database.PgxExt, userID string, groupID string) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM group_participants
			WHERE
				user_id = $1
				AND group_id = $2
		)
	`

	var exists bool
	err := db.QueryRow(ctx, query, userID, groupID).Scan(&exists)

	return exists, err
}

func (r *GroupParticipantRepository) Create(ctx context.Context, db database.PgxExt, participant *entity.GroupParticipant) error {
	query := `
		INSERT INTO group_participants (
			group_id,
			user_id,
			display_name,
			participant_type,
			created_by
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5
		)
	`

	_, err := db.Exec(
		ctx,
		query,

		participant.GroupID,
		participant.UserID,
		participant.DisplayName,
		participant.ParticipantType,
		participant.UserID,
	)

	return err
}
