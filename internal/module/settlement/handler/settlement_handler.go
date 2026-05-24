package handler

import (
	"errors"

	expenseConstants "github.com/KejarBahasa/kejarbill-api/internal/module/expense/constants"

	settlementConstants "github.com/KejarBahasa/kejarbill-api/internal/module/settlement/constants"
	"github.com/KejarBahasa/kejarbill-api/internal/module/settlement/dto"
	"github.com/KejarBahasa/kejarbill-api/internal/module/settlement/service"

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
