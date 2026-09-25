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
	groupMemberConstants "github.com/KejarBahasa/kejarbill-api/internal/module/group_member/constants"
	groupMemberRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/group_member/repository"
	groupParticipantRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/group_participant/repository"
	ledgerConstants "github.com/KejarBahasa/kejarbill-api/internal/module/ledger/constants"
	ledgerEntity "github.com/KejarBahasa/kejarbill-api/internal/module/ledger/entity"
	ledgerRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/ledger/repository"
	ledgerServicePkg "github.com/KejarBahasa/kejarbill-api/internal/module/ledger/service"
	settlementRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/settlement/repository"

	"github.com/google/uuid"
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
	settlementRepo       *settlementRepoPkg.SettlementRepository
}

func NewExpenseService(
	db *pgxpool.Pool,

	expenseRepo *repository.ExpenseRepository,

	ledgerRepo *ledgerRepoPkg.LedgerRepository,
	ledgerService *ledgerServicePkg.LedgerService,

	groupRepo *groupRepoPkg.GroupRepository,
	groupMemberRepo *groupMemberRepoPkg.GroupMemberRepository,
	groupParticipantRepo *groupParticipantRepoPkg.GroupParticipantRepository,
	settlementRepo *settlementRepoPkg.SettlementRepository,
) *ExpenseService {
	return &ExpenseService{
		db: db,

		expenseRepo: expenseRepo,

		ledgerRepo:    ledgerRepo,
		ledgerService: ledgerService,

		groupRepo:            groupRepo,
		groupMemberRepo:      groupMemberRepo,
		groupParticipantRepo: groupParticipantRepo,
		settlementRepo:       settlementRepo,
	}
}

func (s *ExpenseService) CreateEqualExpense(ctx context.Context, userID string, req *dto.CreateExpenseEqualRequest) (string, error) {
	if len(req.ParticipantIDs) == 0 {
		return "", expenseConstants.ErrParticipantsRequired
	}

	validationResult, err := s.validateExpenseCreation(
		ctx,
		expenseConstants.SplitMethodEqual,
		userID,
		req.GroupID,
		req.PayerParticipantID,
		req.ParticipantIDs,
		req.ExpenseDate,
	)
	if err != nil {
		return "", err
	}

	baseShare := req.TotalAmount / int64(len(req.ParticipantIDs))
	remainder := req.TotalAmount % int64(len(req.ParticipantIDs))

	participants := make([]entity.CreateExpenseParticipantPayload, 0, len(req.ParticipantIDs))

	for index, participantID := range req.ParticipantIDs {
		shareAmount := baseShare
		if int64(index) < remainder {
			shareAmount++
		}

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
		expenseConstants.SplitMethodCustom,
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

func (s *ExpenseService) CreateItemizedExpense(ctx context.Context, userID string, req *dto.CreateExpenseItemizedRequest) (string, error) {
	participantShareMap := make(map[string]int64)
	items := make([]entity.ExpenseItem, 0, len(req.Items))
	itemParticipants := make([]entity.ExpenseItemParticipant, 0)
	participantIDs := make([]string, 0, len(req.Items))
	var totalAmount int64

	for _, item := range req.Items {
		if utils.HasDuplicateString(item.ParticipantIDs) {
			return "", expenseConstants.ErrDuplicateParticipants
		}

		subtotal := item.Qty * item.UnitPrice
		baseShare := subtotal / int64(len(item.ParticipantIDs))
		remainder := subtotal % int64(len(item.ParticipantIDs))
		itemID := uuid.NewString()

		items = append(items, entity.ExpenseItem{
			ID:        itemID,
			Name:      item.Name,
			Qty:       item.Qty,
			UnitPrice: item.UnitPrice,
			Subtotal:  subtotal,
			Notes:     utils.PtrOrNil(item.Notes),
		})

		for index, participantID := range item.ParticipantIDs {
			shareAmount := baseShare
			if int64(index) < remainder {
				shareAmount++
			}
			participantIDs = append(participantIDs, participantID)
			participantShareMap[participantID] += shareAmount
			itemParticipants = append(itemParticipants, entity.ExpenseItemParticipant{
				ExpenseItemID: itemID,
				ParticipantID: participantID,
				ShareAmount:   shareAmount,
			})
		}

		totalAmount += subtotal
	}

	validationResult, err := s.validateExpenseCreation(
		ctx,
		expenseConstants.SplitMethodItemized,
		userID,
		req.GroupID,
		req.PayerParticipantID,
		participantIDs,
		req.ExpenseDate,
	)
	if err != nil {
		return "", err
	}

	var expenseID string
	err = database.WithTransaction(ctx, s.db, func(tx database.PgxExt) error {
		createdExpenseID, err := s.expenseRepo.CreateExpense(ctx, tx, &entity.Expense{
			GroupID:             req.GroupID,
			Title:               req.Title,
			Description:         utils.PtrOrNil(req.Description),
			Currency:            req.Currency,
			SubtotalAmount:      totalAmount,
			TotalAmount:         totalAmount,
			ExpenseDate:         validationResult.ExpenseDate,
			PaidByParticipantID: req.PayerParticipantID,
			SplitMethod:         expenseConstants.SplitMethodItemized,
			CreatedBy:           userID,
		})
		if err != nil {
			return err
		}

		expenseID = createdExpenseID

		for i := range items {
			items[i].ExpenseID = expenseID
		}

		err = s.expenseRepo.BulkCreateExpenseItems(ctx, tx, items)
		if err != nil {
			return err
		}

		err = s.expenseRepo.BulkCreateExpenseItemParticipants(ctx, tx, itemParticipants)
		if err != nil {
			return err
		}

		expenseParticipants := make([]entity.ExpenseParticipant, 0, len(participantShareMap))
		ledgers := make([]ledgerEntity.AccountLedger, 0, len(participantShareMap))

		for participantID, shareAmount := range participantShareMap {
			expenseParticipants = append(expenseParticipants, entity.ExpenseParticipant{
				ExpenseID:     expenseID,
				ParticipantID: participantID,
				ShareAmount:   shareAmount,
			})

			if participantID == req.PayerParticipantID {
				continue
			}

			ledgers = append(ledgers, ledgerEntity.AccountLedger{
				GroupID:           req.GroupID,
				FromParticipantID: participantID,
				ToParticipantID:   req.PayerParticipantID,
				Amount:            shareAmount,
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
		return nil, expenseConstants.ErrExpenseNotFound
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

	items, err := s.expenseRepo.FindExpenseItemsByExpenseID(ctx, s.db, expenseID)
	if err != nil {
		return nil, err
	}
	itemParticipants, err := s.expenseRepo.FindExpenseItemParticipantsByExpenseID(ctx, s.db, expenseID)
	if err != nil {
		return nil, err
	}
	participantsByItem := make(map[string][]entity.ExpenseItemParticipant)
	for _, participant := range itemParticipants {
		participantsByItem[participant.ExpenseItemID] = append(participantsByItem[participant.ExpenseItemID], participant)
	}
	itemResponses := make([]entity.ExpenseItem, 0, len(items))

	for _, item := range items {
		itemResponses = append(itemResponses, entity.ExpenseItem{
			ID:           item.ID,
			Name:         item.Name,
			Qty:          item.Qty,
			UnitPrice:    item.UnitPrice,
			Subtotal:     item.Subtotal,
			Notes:        item.Notes,
			Participants: participantsByItem[item.ID],
		})
	}

	expense.Participants = participants
	expense.Items = itemResponses

	return expense, nil
}

func (s *ExpenseService) Update(ctx context.Context, requesterUserID string, expenseID string, req *dto.UpdateExpenseRequest) error {
	_, err := time.Parse(time.RFC3339, req.ExpenseDate)
	if err != nil {
		return expenseConstants.ErrInvalidExpenseDate
	}

	participantShareMap := make(map[string]int64)
	participantIDs := make([]string, 0)
	items := make([]entity.ExpenseItem, 0, len(req.Items))
	itemParticipants := make([]entity.ExpenseItemParticipant, 0)
	var totalAmount int64

	switch req.SplitMethod {
	case expenseConstants.SplitMethodEqual:
		if len(req.ParticipantIDs) == 0 {
			return expenseConstants.ErrParticipantsRequired
		}
		if utils.HasDuplicateString(req.ParticipantIDs) {
			return expenseConstants.ErrDuplicateParticipants
		}
		if req.TotalAmount <= 0 {
			return expenseConstants.ErrInvalidTotalAmount
		}
		baseShare := req.TotalAmount / int64(len(req.ParticipantIDs))
		remainder := req.TotalAmount % int64(len(req.ParticipantIDs))
		for index, participantID := range req.ParticipantIDs {
			shareAmount := baseShare
			if int64(index) < remainder {
				shareAmount++
			}
			participantShareMap[participantID] = shareAmount
			participantIDs = append(participantIDs, participantID)
		}
		totalAmount = req.TotalAmount

	case expenseConstants.SplitMethodCustom:
		if len(req.Participants) == 0 {
			return expenseConstants.ErrParticipantsRequired
		}
		for _, participant := range req.Participants {
			if participant.ShareAmount <= 0 {
				return expenseConstants.ErrInvalidTotalAmount
			}
			participantIDs = append(participantIDs, participant.ParticipantID)
			participantShareMap[participant.ParticipantID] += participant.ShareAmount
			totalAmount += participant.ShareAmount
		}
		if utils.HasDuplicateString(participantIDs) {
			return expenseConstants.ErrDuplicateParticipants
		}

	case expenseConstants.SplitMethodItemized:
		if len(req.Items) == 0 {
			return expenseConstants.ErrParticipantsRequired
		}
		for _, item := range req.Items {
			if len(item.ParticipantIDs) == 0 {
				return expenseConstants.ErrParticipantsRequired
			}
			if utils.HasDuplicateString(item.ParticipantIDs) {
				return expenseConstants.ErrDuplicateParticipants
			}
			subtotal := item.Qty * item.UnitPrice
			baseShare := subtotal / int64(len(item.ParticipantIDs))
			remainder := subtotal % int64(len(item.ParticipantIDs))
			itemID := uuid.NewString()
			items = append(items, entity.ExpenseItem{
				ID:        itemID,
				Name:      item.Name,
				Qty:       item.Qty,
				UnitPrice: item.UnitPrice,
				Subtotal:  subtotal,
				Notes:     utils.PtrOrNil(item.Notes),
			})
			for index, participantID := range item.ParticipantIDs {
				shareAmount := baseShare
				if int64(index) < remainder {
					shareAmount++
				}
				participantIDs = append(participantIDs, participantID)
				participantShareMap[participantID] += shareAmount
				itemParticipants = append(itemParticipants, entity.ExpenseItemParticipant{
					ExpenseItemID: itemID,
					ParticipantID: participantID,
					ShareAmount:   shareAmount,
				})
			}
			totalAmount += subtotal
		}

	default:
		return expenseConstants.ErrInvalidTotalAmount
	}

	return database.WithTransaction(ctx, s.db, func(tx database.PgxExt) error {
		current, err := s.expenseRepo.FindForUpdate(ctx, tx, expenseID)
		if err != nil {
			return expenseConstants.ErrExpenseNotFound
		}
		if current.Status != "active" {
			return expenseConstants.ErrExpenseLocked
		}
		if req.Version != current.Version {
			return expenseConstants.ErrExpenseVersionConflict
		}

		role, err := s.groupMemberRepo.FindActiveMemberRole(ctx, tx, current.GroupID, requesterUserID)
		if err != nil {
			return err
		}
		if role == "" {
			return expenseConstants.ErrForbiddenGroupAccess
		}
		if current.CreatedBy != requesterUserID && !groupMemberConstants.CanManage(role) {
			return expenseConstants.ErrExpenseEditForbidden
		}

		validationResult, err := s.validateExpenseCreation(ctx, req.SplitMethod, requesterUserID, current.GroupID, req.PayerParticipantID, participantIDs, req.ExpenseDate)
		if err != nil {
			return err
		}

		oldParticipants, err := s.expenseRepo.FindExpenseParticipants(ctx, tx, expenseID)
		if err != nil {
			return err
		}
		lockedParticipantIDs := []string{current.PaidByParticipantID, req.PayerParticipantID}
		for _, participant := range oldParticipants {
			lockedParticipantIDs = append(lockedParticipantIDs, participant.ParticipantID)
		}
		lockedParticipantIDs = append(lockedParticipantIDs, participantIDs...)
		lockedParticipantIDs = utils.UniqueStrings(lockedParticipantIDs)
		if err := s.groupParticipantRepo.LockByIDsAndGroupID(ctx, tx, current.GroupID, lockedParticipantIDs); err != nil {
			return err
		}

		oldParticipantIDs := make([]string, 0, len(oldParticipants))
		for _, participant := range oldParticipants {
			oldParticipantIDs = append(oldParticipantIDs, participant.ParticipantID)
		}
		oldParticipantIDs = utils.UniqueStrings(append(oldParticipantIDs, current.PaidByParticipantID))
		settled, err := s.settlementRepo.HasCompletedSettlementForPairs(ctx, tx, current.GroupID, current.PaidByParticipantID, oldParticipantIDs)
		if err != nil {
			return err
		}
		if settled {
			return expenseConstants.ErrExpenseLocked
		}

		if err := s.expenseRepo.DeleteExpenseParticipants(ctx, tx, expenseID); err != nil {
			return err
		}
		if err := s.expenseRepo.DeleteExpenseItemParticipants(ctx, tx, expenseID); err != nil {
			return err
		}
		if err := s.expenseRepo.DeleteExpenseItems(ctx, tx, expenseID); err != nil {
			return err
		}
		if err := s.ledgerRepo.DeleteBySourceID(ctx, tx, expenseID); err != nil {
			return err
		}

		if err := s.expenseRepo.UpdateExpense(ctx, tx, &entity.Expense{
			ID:                  expenseID,
			Title:               req.Title,
			Description:         utils.PtrOrNil(req.Description),
			PaidByParticipantID: req.PayerParticipantID,
			Currency:            req.Currency,
			SubtotalAmount:      totalAmount,
			TotalAmount:         totalAmount,
			SplitMethod:         req.SplitMethod,
			ExpenseDate:         validationResult.ExpenseDate,
			Version:             current.Version,
		}); err != nil {
			return err
		}

		for index := range items {
			items[index].ExpenseID = expenseID
		}
		if req.SplitMethod == expenseConstants.SplitMethodItemized {
			if err := s.expenseRepo.BulkCreateExpenseItems(ctx, tx, items); err != nil {
				return err
			}
			if err := s.expenseRepo.BulkCreateExpenseItemParticipants(ctx, tx, itemParticipants); err != nil {
				return err
			}
		}

		expenseParticipants := make([]entity.ExpenseParticipant, 0, len(participantShareMap))
		ledgers := make([]ledgerEntity.AccountLedger, 0, len(participantShareMap))
		for participantID, shareAmount := range participantShareMap {
			expenseParticipants = append(expenseParticipants, entity.ExpenseParticipant{
				ExpenseID:     expenseID,
				ParticipantID: participantID,
				ShareAmount:   shareAmount,
			})
			if participantID != req.PayerParticipantID {
				ledgers = append(ledgers, ledgerEntity.AccountLedger{
					GroupID:           current.GroupID,
					FromParticipantID: participantID,
					ToParticipantID:   req.PayerParticipantID,
					Amount:            shareAmount,
					SourceType:        ledgerConstants.SourceTypeExpense,
					SourceID:          expenseID,
				})
			}
		}
		if err := s.expenseRepo.BulkCreateExpenseParticipants(ctx, tx, expenseParticipants); err != nil {
			return err
		}
		return s.ledgerRepo.BulkCreate(ctx, tx, ledgers)
	})
}

func (s *ExpenseService) DeleteByID(ctx context.Context, requesterUserID string, expenseID string) error {
	return database.WithTransaction(ctx, s.db, func(tx database.PgxExt) error {
		expense, err := s.expenseRepo.FindForUpdate(ctx, tx, expenseID)
		if err != nil {
			return expenseConstants.ErrExpenseNotFound
		}
		if expense.Status != "active" {
			return expenseConstants.ErrExpenseLocked
		}

		role, err := s.groupMemberRepo.FindActiveMemberRole(ctx, tx, expense.GroupID, requesterUserID)
		if err != nil {
			return err
		}
		if role == "" {
			return expenseConstants.ErrForbiddenGroupAccess
		}
		if expense.CreatedBy != requesterUserID && !groupMemberConstants.CanManage(role) {
			return expenseConstants.ErrExpenseEditForbidden
		}

		participants, err := s.expenseRepo.FindExpenseParticipants(ctx, tx, expenseID)
		if err != nil {
			return err
		}
		participantIDs := []string{expense.PaidByParticipantID}
		for _, participant := range participants {
			participantIDs = append(participantIDs, participant.ParticipantID)
		}
		participantIDs = utils.UniqueStrings(participantIDs)
		if err := s.groupParticipantRepo.LockByIDsAndGroupID(ctx, tx, expense.GroupID, participantIDs); err != nil {
			return err
		}
		settled, err := s.settlementRepo.HasCompletedSettlementForPairs(ctx, tx, expense.GroupID, expense.PaidByParticipantID, participantIDs)
		if err != nil {
			return err
		}
		if settled {
			return expenseConstants.ErrExpenseLocked
		}

		err = s.expenseRepo.DeleteExpenseParticipants(ctx, tx, expenseID)
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
