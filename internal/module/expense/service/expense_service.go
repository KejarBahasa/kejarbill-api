package service

import (
	"context"
	"time"

	expenseConstants "github.com/KejarBahasa/kejarbill-api/internal/module/expense/constants"
	"github.com/KejarBahasa/kejarbill-api/internal/module/expense/dto"
	"github.com/KejarBahasa/kejarbill-api/internal/module/expense/entity"
	"github.com/KejarBahasa/kejarbill-api/internal/module/expense/repository"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/database"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/utils"

	ledgerRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/ledger/repository"
	ledgerServicePkg "github.com/KejarBahasa/kejarbill-api/internal/module/ledger/service"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ExpenseService struct {
	db *pgxpool.Pool

	expenseRepo *repository.ExpenseRepository

	ledgerRepo    *ledgerRepoPkg.LedgerRepository
	ledgerService *ledgerServicePkg.LedgerService
}

func NewExpenseService(
	db *pgxpool.Pool,

	expenseRepo *repository.ExpenseRepository,

	ledgerRepo *ledgerRepoPkg.LedgerRepository,
	ledgerService *ledgerServicePkg.LedgerService,
) *ExpenseService {

	return &ExpenseService{
		db: db,

		expenseRepo: expenseRepo,

		ledgerRepo:    ledgerRepo,
		ledgerService: ledgerService,
	}
}

func (s *ExpenseService) CreateExpenseEqual(ctx context.Context, userID string, req *dto.CreateExpenseEqualRequest) (string, error) {
	if len(req.ParticipantIDs) == 0 {
		return "", expenseConstants.ErrParticipantsRequired
	}

	if utils.HasDuplicateString(req.ParticipantIDs) {
		return "", expenseConstants.ErrDuplicateParticipants
	}

	expenseDate, err := time.Parse(time.RFC3339, req.ExpenseDate)
	if err != nil {
		return "", expenseConstants.ErrInvalidExpenseDate
	}

	if expenseDate.IsZero() {
		expenseDate = time.Now()
	}

	shareAmount := req.TotalAmount / int64(len(req.ParticipantIDs))

	var expenseID string
	err = database.WithTransaction(ctx, s.db, func(tx pgx.Tx) error {
		createdExpenseID, err := s.expenseRepo.CreateExpense(ctx, tx, &entity.Expense{
			GroupID:             req.GroupID,
			Title:               req.Title,
			Description:         utils.PtrOrNil(req.Description),
			PaidByParticipantID: req.PayerParticipantID,
			Currency:            req.Currency,
			SubtotalAmount:      req.TotalAmount,
			TotalAmount:         req.TotalAmount,
			SplitMethod:         expenseConstants.SplitMethodEqual,
			ExpenseDate:         expenseDate,
			CreatedBy:           userID,
		})
		if err != nil {
			return err
		}
		expenseID = createdExpenseID

		participants := make([]entity.ExpenseParticipant, 0, len(req.ParticipantIDs))
		for _, participantID := range req.ParticipantIDs {
			participants = append(participants, entity.ExpenseParticipant{
				ExpenseID:     expenseID,
				ParticipantID: participantID,
				ShareAmount:   shareAmount,
			})
		}

		err = s.expenseRepo.BulkCreateExpenseParticipants(ctx, tx, participants)
		if err != nil {
			return err
		}

		// Ledgers
		ledgers := s.ledgerService.BuildExpenseEntries(
			req.GroupID,
			req.PayerParticipantID,
			expenseID,
			req.ParticipantIDs,
			shareAmount,
		)

		err = s.ledgerRepo.BulkCreate(ctx, tx, ledgers)
		if err != nil {
			return err
		}

		return nil
	})

	return expenseID, err
}
