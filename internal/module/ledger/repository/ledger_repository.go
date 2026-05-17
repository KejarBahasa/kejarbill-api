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

func (r *LedgerRepository) BulkCreate(ctx context.Context, tx pgx.Tx, ledgers []entity.AccountLedger) error {
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

	args := make([]any, 0, len(ledgers)*6)
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
