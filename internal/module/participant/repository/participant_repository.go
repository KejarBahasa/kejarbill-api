package repository

import "context"

type ParticipantRepository struct {
}

func NewParticipantRepository() *ParticipantRepository {
	return &ParticipantRepository{}
}

func (r *ParticipantRepository) CountByIDsAndGroupID(ctx context.Context, groupID string, participantIDs []string) (int, error) {
	// query := `
	// 	SELECT COUNT(*)
	// 	FROM participants
	// 	WHERE
	// 		group_id = $1
	// 		AND id = ANY($2)
	// 		AND deleted_at IS NULL
	// `

	var count int
	// err := r.db.QueryRow(ctx, query, groupID, participantIDs).Scan(&count)
	// if err != nil {
	// 	return 0, err
	// }

	return count, nil
}
