package service

import (
	"context"
	"sort"

	expenseEntity "github.com/KejarBahasa/kejarbill-api/internal/module/expense/entity"
	expenseRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/expense/repository"
	"github.com/jackc/pgx/v5/pgxpool"

	groupRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/group/repository"

	groupMemberRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/group_member/repository"

	settlementEntity "github.com/KejarBahasa/kejarbill-api/internal/module/settlement/entity"
	settlementRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/settlement/repository"

	activityEntity "github.com/KejarBahasa/kejarbill-api/internal/module/activity/entity"

	expenseConstants "github.com/KejarBahasa/kejarbill-api/internal/module/expense/constants"
)

type ActivityService struct {
	db *pgxpool.Pool

	expenseRepo *expenseRepoPkg.ExpenseRepository

	settlementRepo *settlementRepoPkg.SettlementRepository

	groupRepo *groupRepoPkg.GroupRepository

	groupMemberRepo *groupMemberRepoPkg.GroupMemberRepository
}

func NewActivityService(
	db *pgxpool.Pool,

	expenseRepo *expenseRepoPkg.ExpenseRepository,

	settlementRepo *settlementRepoPkg.SettlementRepository,

	groupRepo *groupRepoPkg.GroupRepository,

	groupMemberRepo *groupMemberRepoPkg.GroupMemberRepository,
) *ActivityService {

	return &ActivityService{
		db: db,

		expenseRepo: expenseRepo,

		settlementRepo: settlementRepo,

		groupRepo: groupRepo,

		groupMemberRepo: groupMemberRepo,
	}
}

func (s *ActivityService) GetByGroupID(ctx context.Context, requesterUserID string, groupID string) ([]activityEntity.ActivityTimeline, error) {
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

	expenses, err := s.expenseRepo.FindByGroupID(ctx, s.db, groupID, 50, 0)
	if err != nil {
		return nil, err
	}

	settlements, err := s.settlementRepo.FindByGroupID(ctx, s.db, groupID, 50, 0)
	if err != nil {
		return nil, err
	}

	activities := make([]activityEntity.ActivityTimeline, 0)
	activities = append(activities, mapExpenses(expenses)...)
	activities = append(activities, mapSettlements(settlements)...)

	sort.Slice(activities, func(i, j int) bool {
		return activities[i].CreatedAt.After(
			activities[j].CreatedAt,
		)
	})

	return activities, nil
}

func mapExpenses(expenses []expenseEntity.ExpenseTimeline) []activityEntity.ActivityTimeline {
	activities := make([]activityEntity.ActivityTimeline, 0, len(expenses))

	for _, expense := range expenses {
		activities = append(activities, activityEntity.ActivityTimeline{
			Type:      "expense",
			CreatedAt: expense.ExpenseDate,
			Expense: &activityEntity.ActivityExpense{
				ID:               expense.ID,
				Title:            expense.Title,
				TotalAmount:      expense.TotalAmount,
				PayerDisplayName: expense.PayerDisplayName,
			},
		})
	}

	return activities
}

func mapSettlements(settlements []settlementEntity.SettlementTimeline) []activityEntity.ActivityTimeline {
	activities := make([]activityEntity.ActivityTimeline, 0, len(settlements))

	for _, settlement := range settlements {
		activities = append(activities, activityEntity.ActivityTimeline{
			Type:      "settlement",
			CreatedAt: settlement.SettlementDate,
			Settlement: &activityEntity.ActivitySettlement{
				ID:              settlement.ID,
				Amount:          settlement.Amount,
				FromDisplayName: settlement.FromDisplayName,
				ToDisplayName:   settlement.ToDisplayName,
			},
		})
	}

	return activities
}
