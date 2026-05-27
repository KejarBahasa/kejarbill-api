package route

import (
	"github.com/KejarBahasa/kejarbill-api/internal/module/expense/handler"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/middleware"

	"github.com/gofiber/fiber/v3"
)

func ExpenseRoute(
	api fiber.Router,
	authMiddleware *middleware.AuthMiddleware,
	expenseHandler *handler.ExpenseHandler,
) {
	expense := api.Group("/expenses", authMiddleware.Protected)

	expense.Post("/equal", expenseHandler.CreateEqualExpense)
	expense.Post("/custom", expenseHandler.CreateCustomExpense)
	expense.Post("/itemized", expenseHandler.CreateItemizedExpense)

	expense.Get("/:expense_id", expenseHandler.GetDetailByID)
	expense.Delete("/:expense_id", expenseHandler.DeleteByID)
}
