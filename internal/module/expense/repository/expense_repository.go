package repository

import (
	"context"

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
			expense_id,
			name,
			qty,
			unit_price,
			subtotal,
			notes
		)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	batch := &pgx.Batch{}

	for _, item := range items {
		batch.Queue(
			query,

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

		&expense.PayerParticipantID,
		&expense.PayerDisplayName,
	)
	if err != nil {
		return nil, err
	}

	return &expense, nil
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
