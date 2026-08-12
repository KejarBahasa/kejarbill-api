package repository

import (
	"context"

	"github.com/KejarBahasa/kejarbill-api/internal/module/ledger/dto"
	"github.com/KejarBahasa/kejarbill-api/internal/module/ledger/entity"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/database"
)

type LedgerRepository struct {
}

func NewLedgerRepository() *LedgerRepository {
	return &LedgerRepository{}
}

func (r *LedgerRepository) BulkCreate(ctx context.Context, db database.PgxExt, ledgers []entity.AccountLedger) error {
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

	_, err := db.Exec(ctx, query, args...)

	return err
}

func (r *LedgerRepository) GetGroupBalances(ctx context.Context, db database.PgxExt, groupID string) ([]dto.GroupBalanceResponse, error) {
	query := `
		WITH pair_balances AS (
			SELECT
				LEAST(from_participant_id, to_participant_id) AS participant_a_id,
				GREATEST(from_participant_id, to_participant_id) AS participant_b_id,
				SUM(
					CASE
						WHEN from_participant_id < to_participant_id THEN amount
						ELSE -amount
					END
				)::BIGINT AS amount
			FROM account_ledger
			WHERE group_id = $1
			GROUP BY
				LEAST(from_participant_id, to_participant_id),
				GREATEST(from_participant_id, to_participant_id)
		)
		SELECT
			CASE
				WHEN pb.amount > 0 THEN pb.participant_a_id
				ELSE pb.participant_b_id
			END AS from_participant_id,
			fp.display_name,
			CASE
				WHEN pb.amount > 0 THEN pb.participant_b_id
				ELSE pb.participant_a_id
			END AS to_participant_id,
			tp.display_name,
			ABS(pb.amount)::BIGINT AS amount
		FROM pair_balances pb
		JOIN group_participants fp
			ON fp.id = CASE
				WHEN pb.amount > 0 THEN pb.participant_a_id
				ELSE pb.participant_b_id
			END
		JOIN group_participants tp
			ON tp.id = CASE
				WHEN pb.amount > 0 THEN pb.participant_b_id
				ELSE pb.participant_a_id
			END
		WHERE pb.amount <> 0
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

func (r *LedgerRepository) GetOutstandingBalance(ctx context.Context, db database.PgxExt, groupID string, fromParticipantID string, toParticipantID string) (int64, error) {
	query := `
		SELECT
			COALESCE(SUM(balance_amount), 0)::BIGINT
		FROM (
			SELECT
				CASE
					-- Expense:
					-- B -> A = +100k
					WHEN
						from_participant_id = $2
						AND to_participant_id = $3
					THEN amount

					-- Settlement:
					-- A -> B = -40k
					WHEN
						from_participant_id = $3
						AND to_participant_id = $2
					THEN -amount
					ELSE 0
				END AS balance_amount

			FROM account_ledger
			WHERE
				group_id = $1
		) balances
	`

	var outstanding int64
	err := db.QueryRow(
		ctx,
		query,

		groupID,

		fromParticipantID,
		toParticipantID,
	).Scan(&outstanding)

	return outstanding, err
}

func (r *LedgerRepository) LockParticipantPair(ctx context.Context, db database.PgxExt, groupID string, fromParticipantID string, toParticipantID string) error {
	query := `
		SELECT id
		FROM account_ledger
		WHERE group_id = $1
			AND (
				(from_participant_id = $2 AND to_participant_id = $3)
				OR
				(from_participant_id = $3 AND to_participant_id = $2)
			)
		ORDER BY id
		FOR UPDATE
	`

	rows, err := db.Query(ctx, query, groupID, fromParticipantID, toParticipantID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var ledgerID string
		if err := rows.Scan(&ledgerID); err != nil {
			return err
		}
	}

	return rows.Err()
}

func (r *LedgerRepository) DeleteBySourceID(ctx context.Context, db database.PgxExt, sourceID string) error {
	query := `
		DELETE FROM account_ledger
		WHERE source_id = $1 AND source_type = $2
	`

	_, err := db.Exec(ctx, query, sourceID, "expense")

	return err
}
