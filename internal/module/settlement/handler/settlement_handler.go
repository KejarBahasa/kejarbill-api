package handler

import (
	"errors"
	"time"

	expenseConstants "github.com/KejarBahasa/kejarbill-api/internal/module/expense/constants"

	groupDto "github.com/KejarBahasa/kejarbill-api/internal/module/group/dto"

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

	settlementID, err := h.settlementService.Create(c.Context(), userID, params.GroupID, &body)
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

		default:
			return response.Error(c, fiber.StatusInternalServerError, err.Error(), nil)
		}
	}

	return response.Success(c, "settlement created", fiber.Map{"id": settlementID})
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
