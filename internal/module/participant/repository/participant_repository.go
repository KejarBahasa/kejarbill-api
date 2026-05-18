package repository

import (
	"context"

	"github.com/KejarBahasa/kejarbill-api/internal/shared/database"
)

type ParticipantRepository struct {
}

func NewParticipantRepository() *ParticipantRepository {
	return &ParticipantRepository{}
}

func (r *ParticipantRepository) CountByIDsAndGroupID(ctx context.Context, db database.PgxExt, groupID string, participantIDs []string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM participants
		WHERE
			group_id = $1
			AND id = ANY($2)
			AND deleted_at IS NULL
	`

	var count int
	err := db.QueryRow(ctx, query, groupID, participantIDs).Scan(&count)

	return count, err
}

func (r *ParticipantRepository) ExistsByIDAndGroupID(ctx context.Context, db database.PgxExt, groupID string, participantID string) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM participants
			WHERE
				id = $1
				AND group_id = $2
				AND deleted_at IS NULL
		)
	`

	var exists bool
	err := db.QueryRow(ctx, query, participantID, groupID).Scan(&exists)

	return exists, err
}
