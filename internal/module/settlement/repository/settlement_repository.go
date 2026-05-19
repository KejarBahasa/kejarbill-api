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
