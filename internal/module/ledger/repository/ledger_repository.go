package repository

import (
	"context"

	"github.com/KejarBahasa/kejarbill-api/internal/module/ledger/entity"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/database"

	"github.com/jackc/pgx/v5"
)

type LedgerRepository struct {
}

func NewLedgerRepository() *LedgerRepository {
	return &LedgerRepository{}
}

func (r *LedgerRepository) Create(
	ctx context.Context,
	tx pgx.Tx,

	ledger *entity.AccountLedger,
) error {

	query := `
		INSERT INTO account_ledger (
			group_id,
			from_participant_id,
			to_participant_id,
			source_type,
			source_id,
			amount
		)
		VALUES (
			$1,$2,$3,$4,$5,$6
		)
	`

	_, err := tx.Exec(
		ctx,
		query,

		ledger.GroupID,

		ledger.FromParticipantID,
		ledger.ToParticipantID,

		ledger.SourceType,
		ledger.SourceID,

		ledger.Amount,
	)

	return err
}

func (r *LedgerRepository) BulkCreate(
	ctx context.Context,
	tx pgx.Tx,

	ledgers []*entity.AccountLedger,
) error {

	if len(ledgers) == 0 {
		return nil
	}

	query := database.BuildBulkInsertQuery(
		"account_ledger",
		[]string{
			"group_id",
			"from_participant_id",
			"to_participant_id",
			"source_type",
			"source_id",
			"amount",
		},
		len(ledgers),
	)

	args := make([]any, 0)

	for _, ledger := range ledgers {
		args = append(
			args,

			ledger.GroupID,

			ledger.FromParticipantID,
			ledger.ToParticipantID,

			ledger.SourceType,
			ledger.SourceID,

			ledger.Amount,
		)
	}

	_, err := tx.Exec(ctx, query, args...)

	return err
}
