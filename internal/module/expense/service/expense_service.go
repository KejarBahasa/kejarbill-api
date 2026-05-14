package service

import (
	"context"
	"errors"
	"time"

	expenseConstants "github.com/KejarBahasa/kejarbill-api/internal/module/expense/constants"
	"github.com/KejarBahasa/kejarbill-api/internal/module/expense/dto"
	"github.com/KejarBahasa/kejarbill-api/internal/module/expense/entity"
	"github.com/KejarBahasa/kejarbill-api/internal/module/expense/repository"

	ledgerServicePkg "github.com/KejarBahasa/kejarbill-api/internal/module/ledger/service"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ExpenseService struct {
	db *pgxpool.Pool

	expenseRepo *repository.ExpenseRepository

	ledgerService *ledgerServicePkg.LedgerService
}

func NewExpenseService(
	db *pgxpool.Pool,

	expenseRepo *repository.ExpenseRepository,

	ledgerService *ledgerServicePkg.LedgerService,
) *ExpenseService {

	return &ExpenseService{
		db: db,

		expenseRepo: expenseRepo,

		ledgerService: ledgerService,
	}
}

func (s *ExpenseService) CreateExpenseEqual(
	ctx context.Context,
	userID string,
	req *dto.CreateExpenseEqualRequest,
) error {

	if len(req.ParticipantIDs) == 0 {
		return errors.New("participants required")
	}

	expenseDate, err := time.Parse(time.RFC3339, req.ExpenseDate)
	if err != nil {
		return errors.New("invalid expense date")
	}

	shareAmount := req.TotalAmount / float64(len(req.ParticipantIDs))

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	expense := &entity.Expense{
		GroupID:             req.GroupID,
		Title:               req.Title,
		Description:         &req.Description,
		PaidByParticipantID: req.PayerParticipantID,
		Currency:            req.Currency,
		SubtotalAmount:      req.TotalAmount,
		TotalAmount:         req.TotalAmount,
		SplitMethod:         expenseConstants.SplitMethodEqual,
		ExpenseDate:         expenseDate,
		CreatedBy:           userID,
	}

	expenseID, err := s.expenseRepo.CreateExpense(ctx, tx, expense)
	if err != nil {
		return err
	}

	for _, participantID := range req.ParticipantIDs {
		participant := &entity.ExpenseParticipant{
			ExpenseID: expenseID,

			ParticipantID: participantID,

			ShareAmount: shareAmount,
		}

		err := s.expenseRepo.CreateExpenseParticipant(
			ctx,
			tx,
			participant,
		)

		if err != nil {
			return err
		}

		/*
			SKIP SELF DEBT
		*/
		if participantID == req.PayerParticipantID {
			continue
		}

		err = s.ledgerService.CreateExpenseEntry(
			ctx,
			tx,

			req.GroupID,

			participantID,

			req.PayerParticipantID,

			expenseID,

			shareAmount,
		)

		if err != nil {
			return err
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}
