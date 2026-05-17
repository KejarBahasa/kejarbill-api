package route

import (
	"github.com/KejarBahasa/kejarbill-api/internal/module/expense/handler"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/middleware"

	"github.com/gofiber/fiber/v3"
)

func ExpenseRoute(
	api fiber.Router,
	expenseHandler *handler.ExpenseHandler,
	authMiddleware *middleware.AuthMiddleware,
) {
	expense := api.Group("/expenses", authMiddleware.Protected)

	expense.Post("/equal", expenseHandler.CreateExpenseEqual)
}
