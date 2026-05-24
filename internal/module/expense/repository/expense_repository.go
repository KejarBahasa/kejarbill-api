package repository

import (
	"context"

	"github.com/KejarBahasa/kejarbill-api/internal/module/expense/entity"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/database"
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

func (r *ExpenseRepository) FindByGroupID(ctx context.Context, db database.PgxExt, groupID string) ([]entity.ExpenseTimeline, error) {
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
		WHERE e.group_id = $1
		ORDER BY e.expense_date DESC
	`

	rows, err := db.Query(ctx, query, groupID)
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
