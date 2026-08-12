package handler

import (
	"errors"
	"strings"
	"time"

	expenseConstants "github.com/KejarBahasa/kejarbill-api/internal/module/expense/constants"

	groupDto "github.com/KejarBahasa/kejarbill-api/internal/module/group/dto"

	paymentMethodConstants "github.com/KejarBahasa/kejarbill-api/internal/module/payment_method/constants"
	settlementConstants "github.com/KejarBahasa/kejarbill-api/internal/module/settlement/constants"
	"github.com/KejarBahasa/kejarbill-api/internal/module/settlement/dto"
	"github.com/KejarBahasa/kejarbill-api/internal/module/settlement/service"

	sharedDto "github.com/KejarBahasa/kejarbill-api/internal/shared/dto"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/request"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/response"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/security"

	"github.com/gofiber/fiber/v3"
)

type SettlementHandler struct {
	settlementService *service.SettlementService
}

func NewSettlementHandler(
	settlementService *service.SettlementService,
) *SettlementHandler {

	return &SettlementHandler{
		settlementService: settlementService,
	}
}

func (h *SettlementHandler) Create(c fiber.Ctx) error {
	var params dto.CreateSettlementParams
	if err := request.ValidatePathParams(c, &params); err != nil {
		return request.HandleValidationError(c, err)
	}

	var body dto.CreateSettlementBody
	if err := request.ValidateBody(c, &body); err != nil {
		return request.HandleValidationError(c, err)
	}

	userID := security.GetUserID(c)
	idempotencyKey := strings.TrimSpace(c.Get(settlementConstants.IdempotencyKeyHeader))
	if idempotencyKey == "" {
		return response.Error(c, fiber.StatusBadRequest, settlementConstants.ErrIdempotencyKeyRequired.Error(), nil)
	}

	result, err := h.settlementService.Create(c.Context(), userID, params.GroupID, idempotencyKey, &body)
	if err != nil {
		switch {
		case errors.Is(err, expenseConstants.ErrGroupNotFound):
			return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)

		case errors.Is(err, expenseConstants.ErrForbiddenGroupAccess):
			return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)

		case errors.Is(err, settlementConstants.ErrInvalidSettlementParticipants):
			return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)

		case errors.Is(err, settlementConstants.ErrSettlementAmountExceeded):
			return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)

		case errors.Is(err, settlementConstants.ErrIdempotencyKeyConflict):
			return response.Error(c, fiber.StatusConflict, err.Error(), nil)

		case errors.Is(err, settlementConstants.ErrInvalidSettlementPaymentMethod),
			errors.Is(err, settlementConstants.ErrPaymentMethodMustBeEmpty),
			errors.Is(err, settlementConstants.ErrPaymentMethodRequired),
			errors.Is(err, settlementConstants.ErrPaymentMethodNotOwned),
			errors.Is(err, settlementConstants.ErrInvalidPaymentMethodType),
			errors.Is(err, paymentMethodConstants.ErrPaymentMethodNotFound),
			errors.Is(err, paymentMethodConstants.ErrPaymentMethodInactive):
			return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)

		default:
			return response.Error(c, fiber.StatusInternalServerError, err.Error(), nil)
		}
	}

	return c.Status(result.ResponseCode).Type("json").SendString(result.ResponseBody)
}

func (h *SettlementHandler) GetByGroupID(c fiber.Ctx) error {
	var params groupDto.GroupIDParams
	if err := request.ValidatePathParams(c, &params); err != nil {
		return request.HandleValidationError(c, err)
	}

	var query sharedDto.PaginationQuery
	if err := request.ValidateQuery(c, &query); err != nil {
		return request.HandleValidationError(c, err)
	}

	requesterUserID := security.GetUserID(c)

	settlements, err := h.settlementService.GetByGroupID(c.Context(), requesterUserID, params.GroupID, query.Limit, query.Page)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	result := make([]dto.SettlementTimelineResponse, 0, len(settlements.Settlements))

	for _, settlement := range settlements.Settlements {
		result = append(result, dto.SettlementTimelineResponse{
			ID:             settlement.ID,
			Amount:         settlement.Amount,
			SettlementDate: settlement.SettlementDate.Format(time.RFC3339),
			FromParticipant: dto.SettlementParticipantResponse{
				ParticipantID: settlement.FromParticipantID,
				DisplayName:   settlement.FromDisplayName,
			},
			ToParticipant: dto.SettlementParticipantResponse{
				ParticipantID: settlement.ToParticipantID,
				DisplayName:   settlement.ToDisplayName,
			},
		})
	}

	return response.SuccessWithMeta(
		c,
		"settlements fetched",
		fiber.Map{
			"settlements": result,
		},
		&response.Meta{
			Pagination: &response.PaginationMeta{
				Page:       settlements.Page,
				Limit:      settlements.Limit,
				TotalItems: settlements.TotalItems,
				TotalPages: settlements.TotalPages,
			},
		},
	)
}

func (h *SettlementHandler) GetDetail(c fiber.Ctx) error {
	var params dto.SettlementIDParams
	if err := request.ValidatePathParams(c, &params); err != nil {
		return request.HandleValidationError(c, err)
	}

	requesterUserID := security.GetUserID(c)

	result, err := h.settlementService.GetDetail(c.Context(), requesterUserID, params.SettlementID)
	if err != nil {
		switch {
		case errors.Is(err, settlementConstants.ErrSettlementNotFound):
			return response.Error(c, fiber.StatusNotFound, err.Error(), nil)

		case errors.Is(err, expenseConstants.ErrForbiddenGroupAccess):
			return response.Error(c, fiber.StatusForbidden, err.Error(), nil)

		default:
			return response.Error(c, fiber.StatusInternalServerError, err.Error(), nil)
		}
	}

	return response.Success(c, "settlement detail fetched", result)
}
