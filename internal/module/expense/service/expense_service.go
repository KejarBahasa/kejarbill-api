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

	groupRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/group/repository"
	groupMemberRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/group_member/repository"
	groupParticipantRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/group_participant/repository"
	ledgerRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/ledger/repository"
	ledgerServicePkg "github.com/KejarBahasa/kejarbill-api/internal/module/ledger/service"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ExpenseService struct {
	db *pgxpool.Pool

	expenseRepo *repository.ExpenseRepository

	ledgerRepo    *ledgerRepoPkg.LedgerRepository
	ledgerService *ledgerServicePkg.LedgerService

	groupRepo            *groupRepoPkg.GroupRepository
	groupMemberRepo      *groupMemberRepoPkg.GroupMemberRepository
	groupParticipantRepo *groupParticipantRepoPkg.GroupParticipantRepository
}

func NewExpenseService(
	db *pgxpool.Pool,

	expenseRepo *repository.ExpenseRepository,

	ledgerRepo *ledgerRepoPkg.LedgerRepository,
	ledgerService *ledgerServicePkg.LedgerService,

	groupRepo *groupRepoPkg.GroupRepository,
	groupMemberRepo *groupMemberRepoPkg.GroupMemberRepository,
	groupParticipantRepo *groupParticipantRepoPkg.GroupParticipantRepository,
) *ExpenseService {
	return &ExpenseService{
		db: db,

		expenseRepo: expenseRepo,

		ledgerRepo:    ledgerRepo,
		ledgerService: ledgerService,

		groupRepo:            groupRepo,
		groupMemberRepo:      groupMemberRepo,
		groupParticipantRepo: groupParticipantRepo,
	}
}

func (s *ExpenseService) CreateExpenseEqual(ctx context.Context, userID string, req *dto.CreateExpenseEqualRequest) (string, error) {
	if len(req.ParticipantIDs) == 0 {
		return "", expenseConstants.ErrParticipantsRequired
	}

	groupExists, err := s.groupRepo.ExistsByID(ctx, s.db, req.GroupID)
	if err != nil {
		return "", err
	}
	if !groupExists {
		return "", expenseConstants.ErrGroupNotFound
	}

	hasAccess, err := s.groupMemberRepo.ExistsActiveMember(ctx, s.db, req.GroupID, userID)
	if err != nil {
		return "", err
	}
	if !hasAccess {
		return "", expenseConstants.ErrForbiddenGroupAccess
	}

	payerExists := false
	for _, participantID := range req.ParticipantIDs {
		if participantID == req.PayerParticipantID {
			payerExists = true
			break
		}
	}
	if !payerExists {
		return "", expenseConstants.ErrPayerNotIncludedInParticipants
	}

	payerExistsInGroup, err := s.groupParticipantRepo.ExistsByIDAndGroupID(ctx, s.db, req.GroupID, req.PayerParticipantID)
	if err != nil {
		return "", err
	}
	if !payerExistsInGroup {
		return "", expenseConstants.ErrPayerParticipantNotInGroup
	}

	if utils.HasDuplicateString(req.ParticipantIDs) {
		return "", expenseConstants.ErrDuplicateParticipants
	}

	participantCount, err := s.groupParticipantRepo.CountByIDsAndGroupID(ctx, s.db, req.GroupID, req.ParticipantIDs)
	if err != nil {
		return "", err
	}
	if participantCount != len(req.ParticipantIDs) {
		return "", expenseConstants.ErrParticipantNotInGroup
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
	err = database.WithTransaction(ctx, s.db, func(tx database.PgxExt) error {
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

func (s *ExpenseService) GetByGroupID(ctx context.Context, requesterUserID string, groupID string, page int, limit int) (*entity.PaginatedExpenseTimeline, error) {
	groupExists, err := s.groupRepo.ExistsByID(ctx, s.db, groupID)
	if err != nil {
		return nil, err
	}
	if !groupExists {
		return nil, expenseConstants.ErrGroupNotFound
	}

	hasAccess, err := s.groupMemberRepo.ExistsActiveMember(ctx, s.db, groupID, requesterUserID)
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, expenseConstants.ErrForbiddenGroupAccess
	}

	page, limit = utils.NormalizePagination(page, limit)
	offset := utils.CalculateOffset(page, limit)
	totalItems, err := s.expenseRepo.CountByGroupID(ctx, s.db, groupID)
	if err != nil {
		return nil, err
	}

	expenses, err := s.expenseRepo.FindByGroupID(ctx, s.db, groupID, limit, offset)
	if err != nil {
		return nil, err
	}

	totalPages := utils.CalculateTotalPages(totalItems, limit)

	return &entity.PaginatedExpenseTimeline{
		Expenses:   expenses,
		Page:       page,
		Limit:      limit,
		TotalItems: totalItems,
		TotalPages: totalPages,
	}, nil
}

func (s *ExpenseService) GetDetailByID(ctx context.Context, requesterUserID string, expenseID string) (*entity.ExpenseDetail, error) {
	expense, err := s.expenseRepo.FindDetailByID(ctx, s.db, expenseID)
	if err != nil {
		return nil, err
	}

	hasAccess, err := s.groupMemberRepo.ExistsActiveMember(ctx, s.db, expense.GroupID, requesterUserID)
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, expenseConstants.ErrForbiddenGroupAccess
	}

	participants, err := s.expenseRepo.FindExpenseParticipants(ctx, s.db, expenseID)
	if err != nil {
		return nil, err
	}

	expense.Participants = participants

	return expense, nil
}

func (s *ExpenseService) DeleteByID(ctx context.Context, requesterUserID string, expenseID string) error {
	expense, err := s.expenseRepo.FindDetailByID(ctx, s.db, expenseID)
	if err != nil {
		return expenseConstants.ErrExpenseNotFound
	}

	hasAccess, err := s.groupMemberRepo.ExistsActiveMember(ctx, s.db, expense.GroupID, requesterUserID)
	if err != nil {
		return err
	}
	if !hasAccess {
		return expenseConstants.ErrForbiddenGroupAccess
	}

	return database.WithTransaction(ctx, s.db, func(tx database.PgxExt) error {
		err := s.expenseRepo.DeleteExpenseParticipants(ctx, tx, expenseID)
		if err != nil {
			return err
		}

		err = s.ledgerRepo.DeleteBySourceID(ctx, tx, expenseID)
		if err != nil {
			return err
		}

		err = s.expenseRepo.SoftDeleteExpense(ctx, tx, expenseID)
		if err != nil {
			return err
		}

		return nil
	})
}
