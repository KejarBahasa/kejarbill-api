package repository

import (
	"context"
	"errors"

	"github.com/KejarBahasa/kejarbill-api/internal/module/expense/constants"
	"github.com/KejarBahasa/kejarbill-api/internal/module/expense/entity"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/database"
	"github.com/jackc/pgx/v5"
)

type ExpenseRepository struct {
}

func NewExpenseRepository() *ExpenseRepository {
	return &ExpenseRepository{}
}

func (r *ExpenseRepository) CreateExpense(ctx context.Context, db database.PgxExt, expense *entity.Expense) (string, error) {
	query := `
		INSERT INTO expenses (
			group_id,
			title,
			description,
			paid_by_participant_id,
			currency,
			subtotal_amount,
			total_amount,
			split_method,
			expense_date,
			created_by
		)
		VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10
		)
		RETURNING id
	`

	var expenseID string
	err := db.QueryRow(
		ctx,
		query,

		expense.GroupID,
		expense.Title,
		expense.Description,
		expense.PaidByParticipantID,
		expense.Currency,

		expense.SubtotalAmount,
		expense.TotalAmount,

		expense.SplitMethod,

		expense.ExpenseDate,

		expense.CreatedBy,
	).Scan(&expenseID)

	return expenseID, err
}

func (r *ExpenseRepository) BulkCreateExpenseParticipants(ctx context.Context, db database.PgxExt, participants []entity.ExpenseParticipant) error {
	if len(participants) == 0 {
		return nil
	}

	query := database.BuildBulkInsertQuery(
		"expense_participants",
		[]string{
			"expense_id",
			"participant_id",
			"share_amount",
		},
		len(participants),
	)

	args := make([]any, 0, len(participants)*3)
	for _, participant := range participants {
		args = append(
			args,
			participant.ExpenseID,
			participant.ParticipantID,
			participant.ShareAmount,
		)
	}

	_, err := db.Exec(ctx, query, args...)

	return err
}

func (r *ExpenseRepository) BulkCreateExpenseItems(ctx context.Context, db database.PgxExt, items []entity.ExpenseItem) error {
	if len(items) == 0 {
		return nil
	}

	query := `
		INSERT INTO expense_items (
			id,
			expense_id,
			name,
			qty,
			unit_price,
			subtotal,
			notes
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	batch := &pgx.Batch{}

	for _, item := range items {
		batch.Queue(
			query,

			item.ID,
			item.ExpenseID,
			item.Name,
			item.Qty,
			item.UnitPrice,
			item.Subtotal,
			item.Notes,
		)
	}

	results := db.SendBatch(ctx, batch)

	defer results.Close()

	for range items {
		_, err := results.Exec()
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *ExpenseRepository) BulkCreateExpenseItemParticipants(ctx context.Context, db database.PgxExt, participants []entity.ExpenseItemParticipant) error {
	if len(participants) == 0 {
		return nil
	}

	query := database.BuildBulkInsertQuery(
		"expense_item_participants",
		[]string{"expense_item_id", "participant_id", "share_amount"},
		len(participants),
	)
	args := make([]any, 0, len(participants)*3)
	for _, participant := range participants {
		args = append(args, participant.ExpenseItemID, participant.ParticipantID, participant.ShareAmount)
	}

	_, err := db.Exec(ctx, query, args...)
	return err
}

func (r *ExpenseRepository) GetGroupExpenseTotals(ctx context.Context, db database.PgxExt, groupID string, payerParticipantID any) (groupTotal int64, payerTotal int64, err error) {
	query := `
		SELECT
			COALESCE(SUM(total_amount), 0)::BIGINT,
			COALESCE(SUM(CASE WHEN paid_by_participant_id = $2 THEN total_amount ELSE 0 END), 0)::BIGINT
		FROM expenses
		WHERE group_id = $1
			AND deleted_at IS NULL
	`

	err = db.QueryRow(ctx, query, groupID, payerParticipantID).Scan(&groupTotal, &payerTotal)

	return groupTotal, payerTotal, err
}

func (r *ExpenseRepository) FindMyDebtExpenses(ctx context.Context, db database.PgxExt, groupID string, participantID string, limit int) ([]entity.MyDebtExpense, error) {
	query := `
		WITH pair_balances AS (
			SELECT
				CASE
					WHEN account_ledger.from_participant_id = $2 THEN account_ledger.to_participant_id
					ELSE account_ledger.from_participant_id
				END AS counterparty_id,
				SUM(
					CASE
						WHEN account_ledger.from_participant_id = $2 THEN account_ledger.amount
						ELSE -account_ledger.amount
					END
				)::BIGINT AS net_amount
			FROM account_ledger
			LEFT JOIN expenses ledger_expense
				ON ledger_expense.id = account_ledger.source_id
				AND account_ledger.source_type = 'expense'
			LEFT JOIN settlements ledger_settlement
				ON ledger_settlement.id = account_ledger.source_id
				AND account_ledger.source_type = 'settlement'
			WHERE account_ledger.group_id = $1
				AND (
					account_ledger.from_participant_id = $2
					OR account_ledger.to_participant_id = $2
				)
				AND (
					(account_ledger.source_type = 'expense'
						AND ledger_expense.status = 'active'
						AND ledger_expense.deleted_at IS NULL)
					OR (account_ledger.source_type = 'settlement'
						AND ledger_settlement.status = 'completed')
					OR account_ledger.source_type = 'adjustment'
				)
			GROUP BY counterparty_id
		)
		SELECT
			e.id,
			e.title,
			e.expense_date,
			e.created_at,
			e.paid_by_participant_id,
			payer.display_name,
			ep.share_amount,
			pb.net_amount
		FROM expenses e
		INNER JOIN expense_participants ep
			ON ep.expense_id = e.id
		INNER JOIN group_participants payer
			ON payer.id = e.paid_by_participant_id
		INNER JOIN pair_balances pb
			ON pb.counterparty_id = e.paid_by_participant_id
		WHERE e.group_id = $1
			AND e.status = 'active'
			AND e.deleted_at IS NULL
			AND ep.participant_id = $2
			AND e.paid_by_participant_id <> $2
			AND pb.net_amount > 0
		ORDER BY
			e.paid_by_participant_id,
			e.expense_date ASC,
			e.created_at ASC,
			e.id ASC
		LIMIT $3
	`

	rows, err := db.Query(ctx, query, groupID, participantID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	debts := make([]entity.MyDebtExpense, 0)
	for rows.Next() {
		var debt entity.MyDebtExpense
		if err := rows.Scan(
			&debt.ID,
			&debt.Title,
			&debt.ExpenseDate,
			&debt.CreatedAt,
			&debt.ToParticipantID,
			&debt.ToDisplayName,
			&debt.ShareAmount,
			&debt.NetPairAmount,
		); err != nil {
			return nil, err
		}
		debts = append(debts, debt)
	}

	return debts, rows.Err()
}

func (r *ExpenseRepository) CountByGroupID(ctx context.Context, db database.PgxExt, groupID string) (int64, error) {
	query := `
		SELECT COUNT(*)
		FROM expenses e
		WHERE e.group_id = $1 AND e.deleted_at IS NULL
	`

	var total int64
	err := db.QueryRow(ctx, query, groupID).Scan(&total)

	return total, err
}

func (r *ExpenseRepository) FindByGroupID(ctx context.Context, db database.PgxExt, groupID string, limit int, offset int) ([]entity.ExpenseTimeline, error) {
	query := `
		SELECT
			e.id,
			e.title,
			e.description,
			e.currency,
			e.total_amount,
			e.expense_date,
			p.id,
			p.display_name
		FROM expenses e
		INNER JOIN group_participants p
			ON p.id = e.paid_by_participant_id
		WHERE e.group_id = $1 AND e.deleted_at IS NULL
		ORDER BY e.expense_date DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := db.Query(ctx, query, groupID, limit, offset)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	expenses := make([]entity.ExpenseTimeline, 0)

	for rows.Next() {
		var expense entity.ExpenseTimeline
		err := rows.Scan(
			&expense.ID,
			&expense.Title,
			&expense.Description,
			&expense.Currency,
			&expense.TotalAmount,
			&expense.ExpenseDate,
			&expense.PayerParticipantID,
			&expense.PayerDisplayName,
		)
		if err != nil {
			return nil, err
		}

		expenses = append(expenses, expense)
	}

	return expenses, nil
}

func (r *ExpenseRepository) FindDetailByID(ctx context.Context, db database.PgxExt, expenseID string) (*entity.ExpenseDetail, error) {
	query := `
		SELECT
			e.id,
			e.group_id,
			e.title,
			e.description,
			e.currency,
			e.total_amount,
			e.expense_date,
			e.version,

			p.id,
			p.display_name
		FROM expenses e
		INNER JOIN group_participants p
			ON p.id = e.paid_by_participant_id
		WHERE e.id = $1 AND e.deleted_at IS NULL
	`

	var expense entity.ExpenseDetail

	err := db.QueryRow(ctx, query, expenseID).Scan(
		&expense.ID,
		&expense.GroupID,
		&expense.Title,
		&expense.Description,
		&expense.Currency,
		&expense.TotalAmount,
		&expense.ExpenseDate,
		&expense.Version,

		&expense.PayerParticipantID,
		&expense.PayerDisplayName,
	)
	if err != nil {
		return nil, err
	}

	return &expense, nil
}

func (r *ExpenseRepository) FindForUpdate(ctx context.Context, db database.PgxExt, expenseID string) (*entity.Expense, error) {
	query := `
		SELECT id, group_id, title, description, paid_by_participant_id, currency,
			total_amount, split_method, expense_date, status, created_by, version
		FROM expenses
		WHERE id = $1 AND deleted_at IS NULL
		FOR UPDATE
	`

	var expense entity.Expense
	err := db.QueryRow(ctx, query, expenseID).Scan(
		&expense.ID,
		&expense.GroupID,
		&expense.Title,
		&expense.Description,
		&expense.PaidByParticipantID,
		&expense.Currency,
		&expense.TotalAmount,
		&expense.SplitMethod,
		&expense.ExpenseDate,
		&expense.Status,
		&expense.CreatedBy,
		&expense.Version,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, constants.ErrExpenseNotFound
		}
		return nil, err
	}
	return &expense, nil
}

func (r *ExpenseRepository) UpdateExpense(ctx context.Context, db database.PgxExt, expense *entity.Expense) error {
	query := `
		UPDATE expenses
		SET title = $2, description = $3, paid_by_participant_id = $4,
			currency = $5, subtotal_amount = $6, total_amount = $7,
			split_method = $8, expense_date = $9, version = version + 1,
			updated_at = NOW()
		WHERE id = $1 AND version = $10 AND deleted_at IS NULL
	`
	commandTag, err := db.Exec(ctx, query,
		expense.ID,
		expense.Title,
		expense.Description,
		expense.PaidByParticipantID,
		expense.Currency,
		expense.SubtotalAmount,
		expense.TotalAmount,
		expense.SplitMethod,
		expense.ExpenseDate,
		expense.Version,
	)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return constants.ErrExpenseNotFound
	}
	return nil
}

func (r *ExpenseRepository) DeleteExpenseItemParticipants(ctx context.Context, db database.PgxExt, expenseID string) error {
	_, err := db.Exec(ctx, `
		DELETE FROM expense_item_participants
		WHERE expense_item_id IN (SELECT id FROM expense_items WHERE expense_id = $1)
	`, expenseID)
	return err
}

func (r *ExpenseRepository) DeleteExpenseItems(ctx context.Context, db database.PgxExt, expenseID string) error {
	_, err := db.Exec(ctx, `DELETE FROM expense_items WHERE expense_id = $1`, expenseID)
	return err
}

func (r *ExpenseRepository) FindExpenseParticipants(ctx context.Context, db database.PgxExt, expenseID string) ([]entity.ExpenseDetailParticipant, error) {
	query := `
		SELECT
			ep.participant_id,
			p.display_name,
			p.participant_type,
			ep.share_amount
		FROM expense_participants ep
		INNER JOIN group_participants p
			ON p.id = ep.participant_id
		WHERE ep.expense_id = $1
		ORDER BY p.display_name ASC
	`

	rows, err := db.Query(ctx, query, expenseID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	participants := make([]entity.ExpenseDetailParticipant, 0)

	for rows.Next() {
		var participant entity.ExpenseDetailParticipant
		err := rows.Scan(
			&participant.ParticipantID,
			&participant.DisplayName,
			&participant.ParticipantType,
			&participant.ShareAmount,
		)
		if err != nil {
			return nil, err
		}

		participants = append(participants, participant)
	}

	return participants, nil
}

func (r *ExpenseRepository) FindExpenseItemsByExpenseID(ctx context.Context, db database.PgxExt, expenseID string) ([]entity.ExpenseItem, error) {
	query := `
		SELECT
			id,
			expense_id,
			name,
			qty,
			unit_price,
			subtotal,
			notes
		FROM expense_items
		WHERE expense_id = $1
		ORDER BY created_at ASC
	`

	rows, err := db.Query(ctx, query, expenseID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	items := make([]entity.ExpenseItem, 0)

	for rows.Next() {
		var item entity.ExpenseItem
		err := rows.Scan(
			&item.ID,
			&item.ExpenseID,
			&item.Name,
			&item.Qty,
			&item.UnitPrice,
			&item.Subtotal,
			&item.Notes,
		)
		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return items, nil
}

func (r *ExpenseRepository) FindExpenseItemParticipantsByExpenseID(ctx context.Context, db database.PgxExt, expenseID string) ([]entity.ExpenseItemParticipant, error) {
	query := `
		SELECT
			eip.id,
			eip.expense_item_id,
			eip.participant_id,
			p.display_name,
			eip.share_amount
		FROM expense_item_participants eip
		INNER JOIN expense_items ei ON ei.id = eip.expense_item_id
		INNER JOIN group_participants p ON p.id = eip.participant_id
		WHERE ei.expense_id = $1
		ORDER BY eip.expense_item_id, eip.created_at ASC, eip.id ASC
	`

	rows, err := db.Query(ctx, query, expenseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	participants := make([]entity.ExpenseItemParticipant, 0)
	for rows.Next() {
		var participant entity.ExpenseItemParticipant
		if err := rows.Scan(
			&participant.ID,
			&participant.ExpenseItemID,
			&participant.ParticipantID,
			&participant.DisplayName,
			&participant.ShareAmount,
		); err != nil {
			return nil, err
		}
		participants = append(participants, participant)
	}

	return participants, rows.Err()
}

func (r *ExpenseRepository) DeleteExpenseParticipants(ctx context.Context, db database.PgxExt, expenseID string) error {
	query := `
		DELETE FROM expense_participants
		WHERE expense_id = $1
	`

	_, err := db.Exec(ctx, query, expenseID)

	return err
}

func (r *ExpenseRepository) SoftDeleteExpense(ctx context.Context, db database.PgxExt, expenseID string) error {
	query := `
		UPDATE expenses
		SET deleted_at = NOW()
		WHERE id = $1
	`

	_, err := db.Exec(ctx, query, expenseID)

	return err
}
