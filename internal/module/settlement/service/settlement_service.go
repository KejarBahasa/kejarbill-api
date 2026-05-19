package service

import (
	"context"
	"time"

	expenseConstants "github.com/KejarBahasa/kejarbill-api/internal/module/expense/constants"

	groupRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/group/repository"

	ledgerConstants "github.com/KejarBahasa/kejarbill-api/internal/module/ledger/constants"
	ledgerEntity "github.com/KejarBahasa/kejarbill-api/internal/module/ledger/entity"
	ledgerRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/ledger/repository"

	groupParticipantRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/group_participant/repository"

	"github.com/KejarBahasa/kejarbill-api/internal/module/settlement/dto"

	settlementConstants "github.com/KejarBahasa/kejarbill-api/internal/module/settlement/constants"
	settlementEntity "github.com/KejarBahasa/kejarbill-api/internal/module/settlement/entity"
	settlementRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/settlement/repository"

	"github.com/KejarBahasa/kejarbill-api/internal/shared/database"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/utils"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SettlementService struct {
	db *pgxpool.Pool

	settlementRepo *settlementRepoPkg.SettlementRepository

	ledgerRepo *ledgerRepoPkg.LedgerRepository

	groupRepo *groupRepoPkg.GroupRepository

	groupParticipantRepo *groupParticipantRepoPkg.GroupParticipantRepository
}

func NewSettlementService(
	db *pgxpool.Pool,

	settlementRepo *settlementRepoPkg.SettlementRepository,

	ledgerRepo *ledgerRepoPkg.LedgerRepository,

	groupRepo *groupRepoPkg.GroupRepository,

	groupParticipantRepo *groupParticipantRepoPkg.GroupParticipantRepository,
) *SettlementService {

	return &SettlementService{
		db: db,

		settlementRepo: settlementRepo,

		ledgerRepo: ledgerRepo,

		groupRepo: groupRepo,

		groupParticipantRepo: groupParticipantRepo,
	}
}

func (s *SettlementService) Create(ctx context.Context, userID string, groupID string, req *dto.CreateSettlementBody) (string, error) {
	groupExists, err := s.groupRepo.ExistsByID(ctx, s.db, groupID)
	if err != nil {
		return "", err
	}

	if !groupExists {
		return "", expenseConstants.ErrGroupNotFound
	}

	hasAccess, err := s.groupParticipantRepo.ExistsByUserIDAndGroupID(ctx, s.db, userID, groupID)
	if err != nil {
		return "", err
	}

	if !hasAccess {
		return "", expenseConstants.ErrForbiddenGroupAccess
	}

	if req.FromParticipantID == req.ToParticipantID {
		return "", settlementConstants.ErrInvalidSettlementParticipants
	}

	outstandingBalance, err := s.ledgerRepo.GetOutstandingBalance(ctx, s.db, groupID, req.FromParticipantID, req.ToParticipantID)
	if err != nil {
		return "", err
	}

	if outstandingBalance <= 0 {
		return "", settlementConstants.ErrSettlementAmountExceeded
	}

	if req.Amount > outstandingBalance {
		return "", settlementConstants.ErrSettlementAmountExceeded
	}

	paidAt, err := time.Parse(time.RFC3339, req.PaidAt)
	if err != nil {
		return "", err
	}

	var settlementID string
	err = database.WithTransaction(ctx, s.db, func(tx pgx.Tx) error {
		createdSettlementID, err := s.settlementRepo.Create(ctx, tx, &settlementEntity.Settlement{
			GroupID:           groupID,
			FromParticipantID: req.FromParticipantID,
			ToParticipantID:   req.ToParticipantID,
			Amount:            req.Amount,
			Status:            settlementConstants.StatusCompleted,
			Notes:             utils.PtrOrNil(req.Notes),
			PaidAt:            paidAt,
			CreatedBy:         userID,
		})
		if err != nil {
			return err
		}

		settlementID = createdSettlementID

		/*
			REVERSE LEDGER

			Expense:
			B -> A = 100k

			Settlement:
			A -> B = 40k
		*/
		err = s.ledgerRepo.BulkCreate(ctx, tx, []ledgerEntity.AccountLedger{
			{
				GroupID:           groupID,
				FromParticipantID: req.ToParticipantID,
				ToParticipantID:   req.FromParticipantID,
				SourceType:        ledgerConstants.SourceTypeSettlement,
				SourceID:          settlementID,
				Amount:            req.Amount,
			},
		})
		if err != nil {
			return err
		}

		return nil
	})

	return settlementID, err
}
