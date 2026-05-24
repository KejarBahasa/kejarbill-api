package handler

import (
	"errors"

	expenseConstants "github.com/KejarBahasa/kejarbill-api/internal/module/expense/constants"

	ledgerDto "github.com/KejarBahasa/kejarbill-api/internal/module/ledger/dto"
	ledgerServicePkg "github.com/KejarBahasa/kejarbill-api/internal/module/ledger/service"

	"github.com/KejarBahasa/kejarbill-api/internal/shared/request"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/response"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/security"

	"github.com/gofiber/fiber/v3"
)

type BalanceHandler struct {
	balanceService *ledgerServicePkg.BalanceService
}

func NewBalanceHandler(
	balanceService *ledgerServicePkg.BalanceService,
) *BalanceHandler {
	return &BalanceHandler{
		balanceService: balanceService,
	}
}

func (h *BalanceHandler) GetGroupBalances(c fiber.Ctx) error {
	var params ledgerDto.GetGroupBalancesParams
	if err := request.ValidatePathParams(c, &params); err != nil {
		return request.HandleValidationError(c, err)
	}

	userID := security.GetUserID(c)

	balances, err := h.balanceService.GetGroupBalances(c.Context(), userID, params.GroupID)
	if err != nil {
		switch {
		case errors.Is(err, expenseConstants.ErrGroupNotFound):
			return response.Error(c, fiber.StatusNotFound, err.Error(), nil)

		case errors.Is(err, expenseConstants.ErrForbiddenGroupAccess):
			return response.Error(c, fiber.StatusForbidden, err.Error(), nil)

		default:
			return response.Error(c, fiber.StatusInternalServerError, "internal server error", nil)
		}
	}

	return response.Success(c, "success", balances)
}
