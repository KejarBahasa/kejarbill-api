package repository

import (
	"context"
	"errors"

	"github.com/KejarBahasa/kejarbill-api/internal/module/group_participant/entity"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/database"
	"github.com/jackc/pgx/v5"
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

func (r *GroupParticipantRepository) ClaimGuest(ctx context.Context, db database.PgxExt, participantID string, groupID string, userID string, displayName string) (int64, error) {
	query := `
		UPDATE group_participants
		SET
			user_id = $3,
			participant_type = 'registered',
			display_name = $4,
			claimed_at = NOW(),
			updated_at = NOW()
		WHERE
			id = $1
			AND group_id = $2
			AND participant_type = 'guest'
			AND user_id IS NULL
	`

	tag, err := db.Exec(ctx, query, participantID, groupID, userID, displayName)

	return tag.RowsAffected(), err
}

func (r *GroupParticipantRepository) FindByIDAndGroupID(ctx context.Context, db database.PgxExt, participantID string, groupID string) (*entity.GroupParticipant, error) {
	query := `
		SELECT
			id,
			group_id,
			user_id,
			participant_type
		FROM group_participants
		WHERE id = $1
			AND group_id = $2
	`

	var participant entity.GroupParticipant
	err := db.QueryRow(ctx, query, participantID, groupID).Scan(
		&participant.ID,
		&participant.GroupID,
		&participant.UserID,
		&participant.ParticipantType,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &participant, nil
}

func (r *GroupParticipantRepository) FindByGroupID(ctx context.Context, db database.PgxExt, groupID string) ([]entity.GroupParticipant, error) {
	query := `
		SELECT
			gp.id,
			gp.group_id,
			gp.user_id,
			gp.participant_type,
			gp.display_name,
			u.username,
			gp.created_by,
			gp.created_at,
			gp.updated_at
		FROM group_participants gp
		LEFT JOIN users u ON u.id = gp.user_id
		WHERE gp.group_id = $1
		ORDER BY gp.created_at ASC
	`

	rows, err := db.Query(ctx, query, groupID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	participants := make([]entity.GroupParticipant, 0)

	for rows.Next() {
		var participant entity.GroupParticipant
		err := rows.Scan(
			&participant.ID,
			&participant.GroupID,
			&participant.UserID,
			&participant.ParticipantType,
			&participant.DisplayName,
			&participant.Username,
			&participant.CreatedBy,
			&participant.CreatedAt,
			&participant.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		participants = append(participants, participant)
	}

	return participants, nil
}

func (r *GroupParticipantRepository) FindByUserIDAndGroupID(ctx context.Context, db database.PgxExt, groupID string, userID string) (*entity.GroupParticipant, error) {
	query := `
		SELECT
			id,
			group_id,
			user_id,
			participant_type
		FROM group_participants
		WHERE group_id = $1
			AND user_id = $2
		LIMIT 1
	`

	var participant entity.GroupParticipant
	err := db.QueryRow(ctx, query, groupID, userID).Scan(
		&participant.ID,
		&participant.GroupID,
		&participant.UserID,
		&participant.ParticipantType,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &participant, nil
}

func (r *GroupParticipantRepository) LockByIDsAndGroupID(ctx context.Context, db database.PgxExt, groupID string, participantIDs []string) error {
	query := `
		SELECT id
		FROM group_participants
		WHERE group_id = $1 AND id = ANY($2)
		ORDER BY id
		FOR UPDATE
	`

	rows, err := db.Query(ctx, query, groupID, participantIDs)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var participantID string
		if err := rows.Scan(&participantID); err != nil {
			return err
		}
	}
	return rows.Err()
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
		participant.CreatedBy,
	)

	return err
}

func (r *GroupParticipantRepository) BulkCreate(ctx context.Context, db database.PgxExt, participants []entity.GroupParticipant) error {
	if len(participants) == 0 {
		return nil
	}

	query := database.BuildBulkInsertQuery(
		"group_participants",
		[]string{
			"group_id",
			"user_id",
			"participant_type",
			"display_name",
			"created_by",
		},
		len(participants),
	)

	args := make([]any, 0)
	for _, participant := range participants {
		args = append(
			args,
			participant.GroupID,
			participant.UserID,
			participant.ParticipantType,
			participant.DisplayName,
			participant.CreatedBy,
		)
	}

	_, err := db.Exec(ctx, query, args...)

	return err
}
