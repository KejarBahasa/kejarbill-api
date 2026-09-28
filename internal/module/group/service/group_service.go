package service

import (
	"context"
	"sort"
	"time"

	expenseConstants "github.com/KejarBahasa/kejarbill-api/internal/module/expense/constants"
	expenseEntity "github.com/KejarBahasa/kejarbill-api/internal/module/expense/entity"
	expenseRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/expense/repository"

	groupConstants "github.com/KejarBahasa/kejarbill-api/internal/module/group/constants"
	groupDto "github.com/KejarBahasa/kejarbill-api/internal/module/group/dto"
	groupEntity "github.com/KejarBahasa/kejarbill-api/internal/module/group/entity"
	groupRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/group/repository"

	ledgerRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/ledger/repository"

	groupMemberConstants "github.com/KejarBahasa/kejarbill-api/internal/module/group_member/constants"
	groupMemberEntity "github.com/KejarBahasa/kejarbill-api/internal/module/group_member/entity"
	groupMemberRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/group_member/repository"

	groupParticipantConstants "github.com/KejarBahasa/kejarbill-api/internal/module/group_participant/constants"
	groupParticipantEntity "github.com/KejarBahasa/kejarbill-api/internal/module/group_participant/entity"
	groupParticipantRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/group_participant/repository"

	userRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/user/repository"

	"github.com/KejarBahasa/kejarbill-api/internal/shared/database"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/utils"

	"github.com/jackc/pgx/v5/pgxpool"
)

type GroupService struct {
	db *pgxpool.Pool

	groupRepo *groupRepoPkg.GroupRepository

	groupMemberRepo *groupMemberRepoPkg.GroupMemberRepository

	groupParticipantRepo *groupParticipantRepoPkg.GroupParticipantRepository

	userRepo *userRepoPkg.UserRepository

	expenseRepo *expenseRepoPkg.ExpenseRepository

	ledgerRepo *ledgerRepoPkg.LedgerRepository
}

func NewGroupService(
	db *pgxpool.Pool,

	groupRepo *groupRepoPkg.GroupRepository,

	groupMemberRepo *groupMemberRepoPkg.GroupMemberRepository,

	groupParticipantRepo *groupParticipantRepoPkg.GroupParticipantRepository,

	userRepo *userRepoPkg.UserRepository,

	expenseRepo *expenseRepoPkg.ExpenseRepository,

	ledgerRepo *ledgerRepoPkg.LedgerRepository,
) *GroupService {

	return &GroupService{
		db: db,

		groupRepo: groupRepo,

		groupMemberRepo: groupMemberRepo,

		groupParticipantRepo: groupParticipantRepo,

		userRepo: userRepo,

		expenseRepo: expenseRepo,

		ledgerRepo: ledgerRepo,
	}
}

func (s *GroupService) Create(ctx context.Context, userID string, name string, description *string) (string, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return "", err
	}

	var groupID string
	err = database.WithTransaction(ctx, s.db, func(tx database.PgxExt) error {
		createdGroupID, err := s.groupRepo.Create(ctx, tx, &groupEntity.Group{
			Name:        name,
			Description: description,
			CreatedBy:   userID,
		})
		if err != nil {
			return err
		}

		groupID = createdGroupID

		err = s.groupMemberRepo.Create(ctx, tx, &groupMemberEntity.GroupMember{
			GroupID: groupID,
			UserID:  userID,
			Role:    groupMemberConstants.RoleOwner,
			Status:  groupMemberConstants.StatusActive,
		})
		if err != nil {
			return err
		}

		err = s.groupParticipantRepo.Create(ctx, tx, &groupParticipantEntity.GroupParticipant{
			GroupID:         groupID,
			UserID:          utils.PtrOrNil(userID),
			DisplayName:     user.Name,
			ParticipantType: groupParticipantConstants.TypeRegistered,
			CreatedBy:       userID,
		})
		if err != nil {
			return err
		}

		return nil
	})

	return groupID, err
}

func (s *GroupService) GetDetailByID(ctx context.Context, requesterUserID string, groupID string) (*groupEntity.GroupDetail, error) {
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

	group, err := s.groupRepo.FindDetailByID(ctx, s.db, groupID)
	if err != nil {
		return nil, err
	}

	return group, nil
}

func (s *GroupService) ListUserGroups(ctx context.Context, userID string) ([]groupEntity.GroupDetail, error) {
	return s.groupRepo.FindDetailsByMember(ctx, s.db, userID)
}

func (s *GroupService) GetSummary(ctx context.Context, userID string, groupID string) (*groupDto.GroupSummaryResponse, error) {
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

	myParticipant, err := s.groupParticipantRepo.FindByUserIDAndGroupID(ctx, s.db, groupID, userID)
	if err != nil {
		return nil, err
	}

	summary := &groupDto.GroupSummaryResponse{}

	var payerParticipantID any
	if myParticipant != nil {
		payerParticipantID = myParticipant.ID
	}

	groupTotal, myTotalPaid, err := s.expenseRepo.GetGroupExpenseTotals(ctx, s.db, groupID, payerParticipantID)
	if err != nil {
		return nil, err
	}
	summary.GroupTotal = groupTotal
	summary.MyTotalPaid = myTotalPaid

	if myParticipant != nil {
		balances, err := s.ledgerRepo.GetGroupBalances(ctx, s.db, groupID)
		if err != nil {
			return nil, err
		}
		for _, balance := range balances {
			switch {
			case balance.FromParticipant.ID == myParticipant.ID:
				summary.MyTotalDebt += balance.Amount
			case balance.ToParticipant.ID == myParticipant.ID:
				summary.MyTotalCredit += balance.Amount
			}
		}
	}

	return summary, nil
}

func (s *GroupService) GetMyDebts(ctx context.Context, userID string, groupID string) (*groupDto.MyDebtsResponse, error) {
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

	participant, err := s.groupParticipantRepo.FindByUserIDAndGroupID(ctx, s.db, groupID, userID)
	if err != nil {
		return nil, err
	}
	if participant == nil {
		return &groupDto.MyDebtsResponse{Debts: []groupDto.MyDebtResponse{}}, nil
	}

	// The endpoint is intentionally non-paginated, but the query is capped to avoid
	// loading an unbounded number of expense rows for a single request.
	const maxDebtExpenseRows = 10000
	expenses, err := s.expenseRepo.FindMyDebtExpenses(ctx, s.db, groupID, participant.ID, maxDebtExpenseRows)
	if err != nil {
		return nil, err
	}

	expensesByParticipant := make(map[string][]expenseEntity.MyDebtExpense)
	for _, expense := range expenses {
		expensesByParticipant[expense.ToParticipantID] = append(expensesByParticipant[expense.ToParticipantID], expense)
	}

	result := &groupDto.MyDebtsResponse{Debts: make([]groupDto.MyDebtResponse, 0, len(expensesByParticipant))}
	for _, debtExpenses := range expensesByParticipant {
		debt := groupDto.MyDebtResponse{
			ToParticipant: groupDto.DebtParticipantResponse{
				ID:          debtExpenses[0].ToParticipantID,
				DisplayName: debtExpenses[0].ToDisplayName,
			},
			Expenses: make([]groupDto.MyDebtExpenseResponse, 0, len(debtExpenses)),
		}

		for _, expense := range debtExpenses {
			debt.TotalAmount += expense.ShareAmount
		}

		remainingReduction := debt.TotalAmount - debtExpenses[0].NetPairAmount
		if remainingReduction < 0 {
			remainingReduction = 0
		}

		for _, expense := range debtExpenses {
			paidAmount := expense.ShareAmount
			if remainingReduction < paidAmount {
				paidAmount = remainingReduction
			}
			remainingReduction -= paidAmount

			remainingAmount := expense.ShareAmount - paidAmount
			status := groupConstants.DebtStatusPaid
			if paidAmount == 0 {
				status = groupConstants.DebtStatusUnpaid
			} else if remainingAmount > 0 {
				status = groupConstants.DebtStatusPartial
			}

			debt.PaidAmount += paidAmount
			debt.RemainingAmount += remainingAmount
			debt.Expenses = append(debt.Expenses, groupDto.MyDebtExpenseResponse{
				ExpenseID:       expense.ID,
				Title:           expense.Title,
				ExpenseDate:     expense.ExpenseDate.Format(time.RFC3339),
				Amount:          expense.ShareAmount,
				PaidAmount:      paidAmount,
				RemainingAmount: remainingAmount,
				Status:          status,
			})
		}

		if debt.RemainingAmount == 0 {
			continue
		}
		debt.Status = groupConstants.DebtStatusPartial
		if debt.PaidAmount == 0 {
			debt.Status = groupConstants.DebtStatusUnpaid
		}

		result.TotalAmount += debt.TotalAmount
		result.PaidAmount += debt.PaidAmount
		result.RemainingAmount += debt.RemainingAmount
		result.Debts = append(result.Debts, debt)
	}

	sort.Slice(result.Debts, func(i, j int) bool {
		return result.Debts[i].RemainingAmount > result.Debts[j].RemainingAmount
	})

	return result, nil
}
