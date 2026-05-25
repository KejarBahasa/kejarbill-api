package repository

import (
	"context"

	"github.com/KejarBahasa/kejarbill-api/internal/module/settlement/entity"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/database"
)

type SettlementRepository struct{}

func NewSettlementRepository() *SettlementRepository {
	return &SettlementRepository{}
}

func (r *SettlementRepository) Create(ctx context.Context, db database.PgxExt, settlement *entity.Settlement) (string, error) {
	query := `
		INSERT INTO settlements (
			group_id,
			from_participant_id,
			to_participant_id,
			amount,
			status,
			notes,
			paid_at,
			created_by
		)
		VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8
		)
		RETURNING id
	`

	var settlementID string
	err := db.QueryRow(
		ctx,
		query,

		settlement.GroupID,
		settlement.FromParticipantID,
		settlement.ToParticipantID,
		settlement.Amount,
		settlement.Status,
		settlement.Notes,
		settlement.PaidAt,
		settlement.CreatedBy,
	).Scan(&settlementID)

	return settlementID, err
}

func (r *SettlementRepository) FindByGroupID(ctx context.Context, db database.PgxExt, groupID string) ([]entity.SettlementTimeline, error) {
	query := `
		SELECT
			s.id,
			s.amount,
			s.currency,
			s.created_at,

			fp.id,
			fp.display_name,

			tp.id,
			tp.display_name
		FROM settlements s
		INNER JOIN group_participants fp
			ON fp.id = s.from_participant_id
		INNER JOIN group_participants tp
			ON tp.id = s.to_participant_id
		WHERE s.group_id = $1
		ORDER BY s.created_at DESC
	`

	rows, err := db.Query(ctx, query, groupID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	settlements := make([]entity.SettlementTimeline, 0)

	for rows.Next() {
		var settlement entity.SettlementTimeline
		err := rows.Scan(
			&settlement.ID,
			&settlement.Amount,
			&settlement.Currency,
			&settlement.SettlementDate,

			&settlement.FromParticipantID,
			&settlement.FromDisplayName,

			&settlement.ToParticipantID,
			&settlement.ToDisplayName,
		)
		if err != nil {
			return nil, err
		}

		settlements = append(settlements, settlement)
	}

	return settlements, nil
}
