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

func (r *ExpenseRepository) CreateExpense(ctx context.Context, tx pgx.Tx, expense *entity.Expense) (string, error) {
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
	err := tx.QueryRow(
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

func (r *ExpenseRepository) BulkCreateExpenseParticipants(ctx context.Context, tx pgx.Tx, participants []entity.ExpenseParticipant) error {
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

	_, err := tx.Exec(ctx, query, args...)

	return err
}
