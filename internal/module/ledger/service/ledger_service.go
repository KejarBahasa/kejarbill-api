package service

import (
	ledgerConstants "github.com/KejarBahasa/kejarbill-api/internal/module/ledger/constants"
	"github.com/KejarBahasa/kejarbill-api/internal/module/ledger/entity"
)

type LedgerService struct {
}

func NewLedgerService() *LedgerService {
	return &LedgerService{}
}

func (s *LedgerService) BuildExpenseEntries(groupID string, payerParticipantID string, expenseID string, participantIDs []string, amount int64) []entity.AccountLedger {
	ledgers := make([]entity.AccountLedger, 0, len(participantIDs)-1)
	for _, participantID := range participantIDs {
		// SKIP SELF DEBT
		if participantID == payerParticipantID {
			continue
		}

		ledgers = append(
			ledgers,
			entity.AccountLedger{
				GroupID:           groupID,
				FromParticipantID: participantID,
				ToParticipantID:   payerParticipantID,
				SourceType:        ledgerConstants.SourceTypeExpense,
				SourceID:          expenseID,
				Amount:            amount,
			},
		)
	}

	return ledgers
}
