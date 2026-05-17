package bootstrap

import (
	authRoutePkg "github.com/KejarBahasa/kejarbill-api/internal/module/auth/route"
	expenseRoutePkg "github.com/KejarBahasa/kejarbill-api/internal/module/expense/route"
	userRoutePkg "github.com/KejarBahasa/kejarbill-api/internal/module/user/route"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/response"

	"github.com/gofiber/fiber/v3"
)

func RegisterRoute(
	app *fiber.App,
	dep *Dependency,
) {
	app.Get("/", func(c fiber.Ctx) error {
		return response.Success(c, "KejarBill API is running", nil)
	})

	v1 := app.Group("/v1")

	// AUTH
	authRoutePkg.AuthRoute(v1, dep.AuthHandler)

	// USER
	userRoutePkg.UserRoute(v1, dep.UserHandler, dep.AuthMiddleware)

	// EXPENSE
	expenseRoutePkg.ExpenseRoute(v1, dep.ExpenseHandler, dep.AuthMiddleware)
}
