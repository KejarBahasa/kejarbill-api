package service

import (
	"context"

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

func (s *ExpenseService) CreateEqualExpense(ctx context.Context, userID string, req *dto.CreateExpenseEqualRequest) (string, error) {
	if len(req.ParticipantIDs) == 0 {
		return "", expenseConstants.ErrParticipantsRequired
	}

	validationResult, err := s.validateExpenseCreation(ctx, userID, req.GroupID, req.PayerParticipantID, req.ParticipantIDs, req.ExpenseDate)
	if err != nil {
		return "", err
	}

	shareAmount := req.TotalAmount / int64(len(req.ParticipantIDs))

	participants := make([]entity.CreateExpenseParticipantPayload, 0, len(req.ParticipantIDs))

	for _, participantID := range req.ParticipantIDs {
		participants = append(participants, entity.CreateExpenseParticipantPayload{
			ParticipantID: participantID,
			ShareAmount:   shareAmount,
		})
	}

	payload := &entity.CreateExpensePayload{
		GroupID:            req.GroupID,
		Title:              req.Title,
		Description:        utils.PtrOrNil(req.Description),
		Currency:           req.Currency,
		ExpenseDate:        validationResult.ExpenseDate,
		PayerParticipantID: req.PayerParticipantID,
		TotalAmount:        req.TotalAmount,
		CreatedBy:          userID,
		SplitMethod:        expenseConstants.SplitMethodEqual,
		Participants:       participants,
	}

	return s.createExpense(ctx, payload)
}

func (s *ExpenseService) CreateCustomExpense(ctx context.Context, userID string, req *dto.CreateExpenseCustomRequest) (string, error) {
	if len(req.Participants) == 0 {
		return "", expenseConstants.ErrParticipantsRequired
	}

	participantIDs := make([]string, 0, len(req.Participants))
	participants := make([]entity.CreateExpenseParticipantPayload, 0, len(req.Participants))
	var totalAmount int64

	for _, participant := range req.Participants {
		participantIDs = append(participantIDs, participant.ParticipantID)

		participants = append(participants, entity.CreateExpenseParticipantPayload{
			ParticipantID: participant.ParticipantID,
			ShareAmount:   participant.ShareAmount,
		})

		totalAmount += participant.ShareAmount
	}

	validationResult, err := s.validateExpenseCreation(
		ctx,
		userID,
		req.GroupID,
		req.PayerParticipantID,
		participantIDs,
		req.ExpenseDate,
	)
	if err != nil {
		return "", err
	}

	payload := &entity.CreateExpensePayload{
		GroupID:            req.GroupID,
		Title:              req.Title,
		Description:        utils.PtrOrNil(req.Description),
		Currency:           req.Currency,
		ExpenseDate:        validationResult.ExpenseDate,
		PayerParticipantID: req.PayerParticipantID,
		TotalAmount:        totalAmount,
		CreatedBy:          userID,
		SplitMethod:        expenseConstants.SplitMethodCustom,
		Participants:       participants,
	}

	return s.createExpense(ctx, payload)
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
