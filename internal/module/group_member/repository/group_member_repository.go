package repository

import (
	"context"

	"github.com/KejarBahasa/kejarbill-api/internal/module/group_member/entity"
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

func (r *GroupMemberRepository) FindActiveMemberUserIDs(ctx context.Context, db database.PgxExt, groupID string, userIDs []string) ([]string, error) {
	query := `
		SELECT user_id
		FROM group_members
		WHERE
			group_id = $1
			AND status = 'active'
			AND user_id = ANY($2)
	`

	rows, err := db.Query(ctx, query, groupID, userIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	memberUserIDs := make([]string, 0)
	for rows.Next() {
		var userID string
		err := rows.Scan(&userID)
		if err != nil {
			return nil, err
		}
		memberUserIDs = append(memberUserIDs, userID)
	}

	return memberUserIDs, nil
}

func (r *GroupMemberRepository) Create(ctx context.Context, db database.PgxExt, member *entity.GroupMember) error {
	query := `
		INSERT INTO group_members (
			group_id,
			user_id,
			role,
			status
		)
		VALUES (
			$1,$2,$3,$4
		)
	`

	_, err := db.Exec(
		ctx,
		query,

		member.GroupID,
		member.UserID,
		member.Role,
		member.Status,
	)

	return err
}

func (r *GroupMemberRepository) BulkCreate(ctx context.Context, db database.PgxExt, members []entity.GroupMember) error {
	if len(members) == 0 {
		return nil
	}

	query := database.BuildBulkInsertQuery(
		"group_members",
		[]string{
			"group_id",
			"user_id",
			"role",
			"status",
		},
		len(members),
	)

	args := make([]any, 0)
	for _, member := range members {
		args = append(
			args,
			member.GroupID,
			member.UserID,
			member.Role,
			member.Status,
		)
	}

	_, err := db.Exec(ctx, query, args...)

	return err
}
