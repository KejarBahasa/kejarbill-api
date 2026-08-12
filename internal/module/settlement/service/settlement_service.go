package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	expenseConstants "github.com/KejarBahasa/kejarbill-api/internal/module/expense/constants"

	groupRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/group/repository"

	ledgerConstants "github.com/KejarBahasa/kejarbill-api/internal/module/ledger/constants"
	ledgerEntity "github.com/KejarBahasa/kejarbill-api/internal/module/ledger/entity"
	ledgerRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/ledger/repository"

	groupMemberRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/group_member/repository"
	groupParticipantRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/group_participant/repository"

	paymentMethodRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/payment_method/repository"

	"github.com/KejarBahasa/kejarbill-api/internal/module/settlement/dto"
	"github.com/KejarBahasa/kejarbill-api/internal/module/settlement/entity"

	settlementConstants "github.com/KejarBahasa/kejarbill-api/internal/module/settlement/constants"
	settlementEntity "github.com/KejarBahasa/kejarbill-api/internal/module/settlement/entity"
	settlementRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/settlement/repository"

	"github.com/KejarBahasa/kejarbill-api/internal/shared/database"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/response"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/utils"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SettlementService struct {
	db *pgxpool.Pool

	settlementRepo *settlementRepoPkg.SettlementRepository

	ledgerRepo *ledgerRepoPkg.LedgerRepository

	groupRepo *groupRepoPkg.GroupRepository

	groupMemberRepo *groupMemberRepoPkg.GroupMemberRepository

	groupParticipantRepo *groupParticipantRepoPkg.GroupParticipantRepository

	paymentMethodRepo *paymentMethodRepoPkg.PaymentMethodRepository
}

func NewSettlementService(
	db *pgxpool.Pool,

	settlementRepo *settlementRepoPkg.SettlementRepository,

	ledgerRepo *ledgerRepoPkg.LedgerRepository,

	groupRepo *groupRepoPkg.GroupRepository,

	groupMemberRepo *groupMemberRepoPkg.GroupMemberRepository,

	groupParticipantRepo *groupParticipantRepoPkg.GroupParticipantRepository,

	paymentMethodRepo *paymentMethodRepoPkg.PaymentMethodRepository,
) *SettlementService {

	return &SettlementService{
		db: db,

		settlementRepo: settlementRepo,

		ledgerRepo: ledgerRepo,

		groupRepo: groupRepo,

		groupMemberRepo: groupMemberRepo,

		groupParticipantRepo: groupParticipantRepo,

		paymentMethodRepo: paymentMethodRepo,
	}
}

type CreateSettlementResult struct {
	ResponseCode int
	ResponseBody string
}

type createSettlementHashPayload struct {
	UserID            string `json:"user_id"`
	GroupID           string `json:"group_id"`
	FromParticipantID string `json:"from_participant_id"`
	ToParticipantID   string `json:"to_participant_id"`
	Amount            int64  `json:"amount"`
	PaymentChannel    string `json:"payment_channel"`
	PaymentMethodID   string `json:"payment_method_id"`
	Notes             string `json:"notes"`
	PaidAt            string `json:"paid_at"`
}

func (s *SettlementService) Create(ctx context.Context, userID string, groupID string, idempotencyKey string, req *dto.CreateSettlementBody) (*CreateSettlementResult, error) {
	requestHash, err := buildCreateSettlementRequestHash(userID, groupID, req)
	if err != nil {
		return nil, err
	}

	var settlementID string
	result := &CreateSettlementResult{}
	err = database.WithTransaction(ctx, s.db, func(tx database.PgxExt) error {
		claimed, err := s.settlementRepo.CreateIdempotencyKey(ctx, tx, &settlementEntity.IdempotencyKey{
			UserID:      userID,
			Key:         idempotencyKey,
			RequestHash: requestHash,
			ExpiredAt:   time.Now().Add(24 * time.Hour),
		})
		if err != nil {
			return err
		}

		if !claimed {
			existingKey, err := s.settlementRepo.FindIdempotencyKeyByKey(ctx, tx, idempotencyKey)
			if err != nil {
				return err
			}
			if existingKey.RequestHash != requestHash {
				return settlementConstants.ErrIdempotencyKeyConflict
			}
			if existingKey.ResponseCode == nil || existingKey.ResponseBody == nil {
				return settlementConstants.ErrIdempotencyResponseUnavailable
			}

			result.ResponseCode = *existingKey.ResponseCode
			result.ResponseBody = *existingKey.ResponseBody

			return nil
		}

		groupExists, err := s.groupRepo.ExistsByID(ctx, tx, groupID)
		if err != nil {
			return err
		}
		if !groupExists {
			return expenseConstants.ErrGroupNotFound
		}

		hasAccess, err := s.groupMemberRepo.ExistsActiveMember(ctx, tx, groupID, userID)
		if err != nil {
			return err
		}
		if !hasAccess {
			return expenseConstants.ErrForbiddenGroupAccess
		}

		if req.FromParticipantID == req.ToParticipantID {
			return settlementConstants.ErrInvalidSettlementParticipants
		}

		toParticipant, err := s.groupParticipantRepo.FindByIDAndGroupID(ctx, tx, req.ToParticipantID, groupID)
		if err != nil {
			return err
		}
		if toParticipant == nil {
			return settlementConstants.ErrInvalidSettlementParticipants
		}

		if err := s.validatePaymentMethod(ctx, tx, *req, toParticipant); err != nil {
			return err
		}

		paidAt, err := time.Parse(time.RFC3339, req.PaidAt)
		if err != nil {
			return err
		}

		if err := s.ledgerRepo.LockParticipantPair(ctx, tx, groupID, req.FromParticipantID, req.ToParticipantID); err != nil {
			return err
		}

		outstandingBalance, err := s.ledgerRepo.GetOutstandingBalance(ctx, tx, groupID, req.FromParticipantID, req.ToParticipantID)
		if err != nil {
			return err
		}

		if outstandingBalance <= 0 || req.Amount > outstandingBalance {
			return settlementConstants.ErrSettlementAmountExceeded
		}

		createdSettlementID, err := s.settlementRepo.Create(ctx, tx, &settlementEntity.Settlement{
			GroupID:           groupID,
			FromParticipantID: req.FromParticipantID,
			ToParticipantID:   req.ToParticipantID,
			PaymentChannel:    req.PaymentChannel,
			PaymentMethodID:   utils.PtrOrNil(req.PaymentMethodID),
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

		responseBody, err := json.Marshal(response.Response[map[string]string]{
			Status:  response.StatusSuccess,
			Message: "settlement created",
			Data:    map[string]string{"id": settlementID},
		})
		if err != nil {
			return err
		}

		if err := s.settlementRepo.SaveIdempotencyResponse(ctx, tx, idempotencyKey, http.StatusOK, string(responseBody)); err != nil {
			return err
		}

		result.ResponseCode = http.StatusOK
		result.ResponseBody = string(responseBody)

		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

func buildCreateSettlementRequestHash(userID string, groupID string, req *dto.CreateSettlementBody) (string, error) {
	payload, err := json.Marshal(createSettlementHashPayload{
		UserID:            userID,
		GroupID:           groupID,
		FromParticipantID: req.FromParticipantID,
		ToParticipantID:   req.ToParticipantID,
		Amount:            req.Amount,
		PaymentChannel:    req.PaymentChannel,
		PaymentMethodID:   req.PaymentMethodID,
		Notes:             req.Notes,
		PaidAt:            req.PaidAt,
	})
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(payload)

	return hex.EncodeToString(hash[:]), nil
}

func (s *SettlementService) GetByGroupID(ctx context.Context, requesterUserID string, groupID string, limit, page int) (*entity.PaginatedSettlementTimeline, error) {
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
	totalItems, err := s.settlementRepo.CountByGroupID(ctx, s.db, groupID)
	if err != nil {
		return nil, err
	}

	settlements, err := s.settlementRepo.FindByGroupID(ctx, s.db, groupID, limit, offset)
	if err != nil {
		return nil, err
	}

	totalPages := utils.CalculateTotalPages(totalItems, limit)

	return &entity.PaginatedSettlementTimeline{
		Settlements: settlements,
		Page:        page,
		Limit:       limit,
		TotalItems:  totalItems,
		TotalPages:  totalPages,
	}, nil
}

func (s *SettlementService) GetDetail(ctx context.Context, requesterUserID string, settlementID string) (*dto.SettlementDetailResponse, error) {
	settlement, err := s.settlementRepo.FindDetailByID(ctx, s.db, settlementID)
	if err != nil {
		return nil, err
	}

	hasAccess, err := s.groupMemberRepo.ExistsActiveMember(ctx, s.db, settlement.GroupID, requesterUserID)
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, expenseConstants.ErrForbiddenGroupAccess
	}

	response := &dto.SettlementDetailResponse{
		ID:             settlement.ID,
		GroupID:        settlement.GroupID,
		Amount:         settlement.Amount,
		PaymentChannel: settlement.PaymentChannel,
		Status:         settlement.Status,
		PaidAt:         settlement.PaidAt.Format(time.RFC3339),
		Notes:          settlement.Notes,
		FromParticipant: dto.SettlementDetailParticipantInfo{
			ParticipantID: settlement.FromParticipantID,
			DisplayName:   settlement.FromDisplayName,
		},
		ToParticipant: dto.SettlementDetailParticipantInfo{
			ParticipantID: settlement.ToParticipantID,
			DisplayName:   settlement.ToDisplayName,
		},
	}

	if settlement.PaymentMethodID != nil {
		response.PaymentMethod = &dto.SettlementDetailPaymentMethodInfo{
			ID:           *settlement.PaymentMethodID,
			MethodType:   utils.DerefString(settlement.MethodType),
			ProviderName: utils.DerefString(settlement.ProviderName),
		}
	}

	return response, nil
}
