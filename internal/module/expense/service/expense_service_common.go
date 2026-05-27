package service

import (
	"context"
	"time"

	expenseConstants "github.com/KejarBahasa/kejarbill-api/internal/module/expense/constants"
	"github.com/KejarBahasa/kejarbill-api/internal/module/expense/entity"
	ledgerConstants "github.com/KejarBahasa/kejarbill-api/internal/module/ledger/constants"
	ledgerEntity "github.com/KejarBahasa/kejarbill-api/internal/module/ledger/entity"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/database"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/utils"
)

func (s *ExpenseService) validateExpenseCreation(ctx context.Context, userID string, groupID string, payerParticipantID string, participantIDs []string, expenseDate string) (*entity.ExpenseValidationResult, error) {
	groupExists, err := s.groupRepo.ExistsByID(ctx, s.db, groupID)
	if err != nil {
		return nil, err
	}
	if !groupExists {
		return nil, expenseConstants.ErrGroupNotFound
	}

	hasAccess, err := s.groupMemberRepo.ExistsActiveMember(ctx, s.db, groupID, userID)
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, expenseConstants.ErrForbiddenGroupAccess
	}

	if len(participantIDs) == 0 {
		return nil, expenseConstants.ErrParticipantsRequired
	}

	if utils.HasDuplicateString(participantIDs) {
		return nil, expenseConstants.ErrDuplicateParticipants
	}

	payerExists := false

	for _, participantID := range participantIDs {
		if participantID == payerParticipantID {
			payerExists = true
			break
		}
	}

	if !payerExists {
		return nil, expenseConstants.ErrPayerNotIncludedInParticipants
	}

	payerExistsInGroup, err := s.groupParticipantRepo.ExistsByIDAndGroupID(ctx, s.db, groupID, payerParticipantID)
	if err != nil {
		return nil, err
	}

	if !payerExistsInGroup {
		return nil, expenseConstants.ErrPayerParticipantNotInGroup
	}

	participantCount, err := s.groupParticipantRepo.CountByIDsAndGroupID(ctx, s.db, groupID, participantIDs)
	if err != nil {
		return nil, err
	}

	if participantCount != len(participantIDs) {
		return nil, expenseConstants.ErrParticipantNotInGroup
	}

	parsedExpenseDate, err := time.Parse(time.RFC3339, expenseDate)
	if err != nil {
		return nil, expenseConstants.ErrInvalidExpenseDate
	}
	if parsedExpenseDate.IsZero() {
		parsedExpenseDate = time.Now()
	}

	return &entity.ExpenseValidationResult{
		ExpenseDate: parsedExpenseDate,
	}, nil
}

func (s *ExpenseService) createExpense(ctx context.Context, payload *entity.CreateExpensePayload) (string, error) {
	var totalAmount int64
	for _, participant := range payload.Participants {
		totalAmount += participant.ShareAmount
	}
	if totalAmount != payload.TotalAmount {
		return "", expenseConstants.ErrInvalidTotalAmount
	}

	var expenseID string
	err := database.WithTransaction(ctx, s.db, func(tx database.PgxExt) error {
		createdExpenseID, err := s.expenseRepo.CreateExpense(ctx, tx, &entity.Expense{
			GroupID:             payload.GroupID,
			Title:               payload.Title,
			Description:         payload.Description,
			Currency:            payload.Currency,
			SubtotalAmount:      payload.TotalAmount,
			TotalAmount:         payload.TotalAmount,
			ExpenseDate:         payload.ExpenseDate,
			PaidByParticipantID: payload.PayerParticipantID,
			SplitMethod:         payload.SplitMethod,
			CreatedBy:           payload.CreatedBy,
		})
		if err != nil {
			return err
		}
		expenseID = createdExpenseID

		expenseParticipants := make([]entity.ExpenseParticipant, 0, len(payload.Participants))
		ledgerCapacity := 0
		if len(payload.Participants) > 1 {
			ledgerCapacity = len(payload.Participants) - 1
		}
		ledgers := make([]ledgerEntity.AccountLedger, 0, ledgerCapacity)

		for _, participant := range payload.Participants {
			expenseParticipants = append(expenseParticipants, entity.ExpenseParticipant{
				ExpenseID:     expenseID,
				ParticipantID: participant.ParticipantID,
				ShareAmount:   participant.ShareAmount,
			})

			if participant.ParticipantID == payload.PayerParticipantID {
				continue
			}

			ledgers = append(ledgers, ledgerEntity.AccountLedger{
				GroupID:           payload.GroupID,
				FromParticipantID: participant.ParticipantID,
				ToParticipantID:   payload.PayerParticipantID,
				Amount:            participant.ShareAmount,
				SourceType:        ledgerConstants.SourceTypeExpense,
				SourceID:          expenseID,
			})
		}

		err = s.expenseRepo.BulkCreateExpenseParticipants(ctx, tx, expenseParticipants)
		if err != nil {
			return err
		}

		if len(ledgers) > 0 {
			err = s.ledgerRepo.BulkCreate(ctx, tx, ledgers)
			if err != nil {
				return err
			}
		}

		return nil
	})

	return expenseID, err
}
