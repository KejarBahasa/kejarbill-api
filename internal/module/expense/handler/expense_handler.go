package handler

import (
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

	if err := request.ValidateBody(c, &req); err != nil {
		return err
	}

	userID := security.GetUserID(c)

	err := h.expenseService.CreateExpenseEqual(c.Context(), userID, &req)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return response.Success(c, "expense created", nil)
}
