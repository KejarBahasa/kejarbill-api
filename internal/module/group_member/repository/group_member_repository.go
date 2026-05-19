package repository

import (
	"context"

	"github.com/KejarBahasa/kejarbill-api/internal/shared/database"
)

type GroupMemberRepository struct {
}

func NewGroupMemberRepository() *GroupMemberRepository {
	return &GroupMemberRepository{}
}

func (r *GroupMemberRepository) ExistsActiveMember(ctx context.Context, db database.PgxExt, groupID string, userID string) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM group_members
			WHERE
				group_id = $1
				AND user_id = $2
				AND status = 'active'
		)
	`

	var exists bool
	err := db.QueryRow(ctx, query, groupID, userID).Scan(&exists)

	return exists, err
}
