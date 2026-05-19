package handler

import (
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
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return response.Success(c, "settlement created", fiber.Map{"id": settlementID})
}
