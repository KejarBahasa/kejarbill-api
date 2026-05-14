package service

import (
	"context"

	ledgerConstants "github.com/KejarBahasa/kejarbill-api/internal/module/ledger/constants"
	"github.com/KejarBahasa/kejarbill-api/internal/module/ledger/entity"
	"github.com/KejarBahasa/kejarbill-api/internal/module/ledger/repository"

	"github.com/jackc/pgx/v5"
)

type LedgerService struct {
	ledgerRepo *repository.LedgerRepository
}

func NewLedgerService(
	ledgerRepo *repository.LedgerRepository,
) *LedgerService {

	return &LedgerService{
		ledgerRepo: ledgerRepo,
	}
}

func (s *LedgerService) CreateExpenseEntry(
	ctx context.Context,
	tx pgx.Tx,

	groupID string,

	fromParticipantID string,
	toParticipantID string,

	expenseID string,

	amount float64,
) error {

	ledger := &entity.AccountLedger{
		GroupID: groupID,

		FromParticipantID: fromParticipantID,

		ToParticipantID: toParticipantID,

		SourceType: ledgerConstants.SourceTypeExpense,

		SourceID: expenseID,

		Amount: amount,
	}

	return s.ledgerRepo.Create(
		ctx,
		tx,
		ledger,
	)
}
