package repository

import (
	"context"

	"github.com/KejarBahasa/kejarbill-api/internal/module/ledger/dto"
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

func (r *LedgerRepository) GetGroupBalances(ctx context.Context, db database.PgxExt, groupID string) ([]dto.GroupBalanceResponse, error) {
	query := `
		SELECT
			al.from_participant_id,
			fp.display_name,
			al.to_participant_id,
			tp.display_name,
			SUM(al.amount)::BIGINT AS amount
		FROM account_ledger al
		JOIN group_participants fp
			ON fp.id = al.from_participant_id
		JOIN group_participants tp
			ON tp.id = al.to_participant_id
		WHERE
			al.group_id = $1
		GROUP BY
			al.from_participant_id,
			fp.display_name,
			al.to_participant_id,
			tp.display_name
		HAVING SUM(al.amount) > 0
		ORDER BY amount DESC
	`

	rows, err := db.Query(
		ctx,
		query,
		groupID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	balances := make([]dto.GroupBalanceResponse, 0)
	for rows.Next() {
		var balance dto.GroupBalanceResponse
		err := rows.Scan(
			&balance.FromParticipant.ID,

			&balance.FromParticipant.DisplayName,
			&balance.ToParticipant.ID,
			&balance.ToParticipant.DisplayName,

			&balance.Amount,
		)
		if err != nil {
			return nil, err
		}

		balances = append(balances, balance)
	}

	return balances, nil
}
