package repository

import (
	"context"

	"github.com/KejarBahasa/kejarbill-api/internal/module/group/entity"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/database"
)

type GroupRepository struct {
}

func NewGroupRepository() *GroupRepository {
	return &GroupRepository{}
}

func (r *GroupRepository) ExistsByID(ctx context.Context, db database.PgxExt, groupID string) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM groups
			WHERE
				id = $1
				AND deleted_at IS NULL
		)
	`

	var exists bool
	err := db.QueryRow(ctx, query, groupID).Scan(&exists)

	return exists, err
}

func (r *GroupRepository) Create(ctx context.Context, db database.PgxExt, group *entity.Group) (string, error) {
	query := `
		INSERT INTO groups (
			name, created_by
		)
		VALUES (
			$1, $2
		)
		RETURNING id
	`

	var groupID string
	err := db.QueryRow(ctx, query, group.Name, group.CreatedBy).Scan(&groupID)

	return groupID, err
}

func (r *GroupRepository) FindDetailByID(ctx context.Context, db database.PgxExt, groupID string) (*entity.GroupDetail, error) {
	query := `
		SELECT
			g.id,
			g.name,
			g.description,
			(
				SELECT COUNT(*)
				FROM group_members gm
				WHERE
					gm.group_id = g.id
					AND gm.status = 'active'
			),
			(
				SELECT COUNT(*)
				FROM group_participants gp
				WHERE gp.group_id = g.id
			),
			(
				SELECT COUNT(*)
				FROM expenses e
				WHERE
					e.group_id = g.id
					AND e.deleted_at IS NULL
			),
			g.created_at
		FROM groups g
		WHERE g.id = $1
	`

	var group entity.GroupDetail
	err := db.QueryRow(ctx, query, groupID).Scan(
		&group.ID,
		&group.Name,
		&group.Description,
		&group.TotalMembers,
		&group.TotalParticipants,
		&group.TotalExpenses,
		&group.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &group, nil
}
