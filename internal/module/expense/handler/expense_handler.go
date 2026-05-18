package handler

import (
	"errors"

	expenseConstants "github.com/KejarBahasa/kejarbill-api/internal/module/expense/constants"
	"github.com/KejarBahasa/kejarbill-api/internal/module/expense/dto"
	"github.com/KejarBahasa/kejarbill-api/internal/module/expense/service"

	"github.com/KejarBahasa/kejarbill-api/internal/shared/request"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/response"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/security"

	"github.com/gofiber/fiber/v3"
)

type ExpenseHandler struct {
	expenseService *service.ExpenseService
}

func NewExpenseHandler(
	expenseService *service.ExpenseService,
) *ExpenseHandler {

	return &ExpenseHandler{
		expenseService: expenseService,
	}
}

func (h *ExpenseHandler) CreateExpenseEqual(c fiber.Ctx) error {
	var req dto.CreateExpenseEqualRequest
	if validationErr := request.ValidateBody(c, &req); validationErr != nil {
		return request.HandleValidationError(c, validationErr)
	}

	userID := security.GetUserID(c)

	expenseID, err := h.expenseService.CreateExpenseEqual(c.Context(), userID, &req)
	if err != nil {
		switch {
		case errors.Is(err, expenseConstants.ErrParticipantsRequired):
			return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)

		case errors.Is(err, expenseConstants.ErrGroupNotFound):
			return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)

		case errors.Is(err, expenseConstants.ErrPayerNotIncludedInParticipants):
			return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)

		case errors.Is(err, expenseConstants.ErrPayerParticipantNotInGroup):
			return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)

		case errors.Is(err, expenseConstants.ErrDuplicateParticipants):
			return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)

		case errors.Is(err, expenseConstants.ErrParticipantNotInGroup):
			return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)

		case errors.Is(err, expenseConstants.ErrInvalidExpenseDate):
			return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)

		default:
			return response.Error(c, fiber.StatusInternalServerError, err.Error(), nil)
		}
	}

	return response.Success(c, "expense created", fiber.Map{"id": expenseID})
}
