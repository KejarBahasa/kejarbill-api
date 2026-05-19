package service

import (
	"context"

	groupRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/group/repository"

	ledgerDto "github.com/KejarBahasa/kejarbill-api/internal/module/ledger/dto"
	ledgerRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/ledger/repository"

	groupParticipantRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/group_participant/repository"

	expenseConstants "github.com/KejarBahasa/kejarbill-api/internal/module/expense/constants"

	"github.com/jackc/pgx/v5/pgxpool"
)

type BalanceService struct {
	db                   *pgxpool.Pool
	ledgerRepo           *ledgerRepoPkg.LedgerRepository
	groupRepo            *groupRepoPkg.GroupRepository
	groupParticipantRepo *groupParticipantRepoPkg.GroupParticipantRepository
}

func NewBalanceService(
	db *pgxpool.Pool,
	ledgerRepo *ledgerRepoPkg.LedgerRepository,
	groupRepo *groupRepoPkg.GroupRepository,
	groupParticipantRepo *groupParticipantRepoPkg.GroupParticipantRepository,
) *BalanceService {
	return &BalanceService{
		db: db,

		ledgerRepo: ledgerRepo,

		groupRepo: groupRepo,

		groupParticipantRepo: groupParticipantRepo,
	}
}

func (s *BalanceService) GetGroupBalances(ctx context.Context, userID, groupID string) ([]ledgerDto.GroupBalanceResponse, error) {
	groupExists, err := s.groupRepo.ExistsByID(ctx, s.db, groupID)
	if err != nil {
		return nil, err
	}
	if !groupExists {
		return nil, expenseConstants.ErrGroupNotFound
	}

	hasAccess, err := s.groupParticipantRepo.ExistsByUserIDAndGroupID(ctx, s.db, userID, groupID)
	if err != nil {
		return nil, expenseConstants.ErrForbiddenGroupAccess
	}
	if !hasAccess {
		return nil, expenseConstants.ErrForbiddenGroupAccess
	}

	return s.ledgerRepo.GetGroupBalances(ctx, s.db, groupID)
}
