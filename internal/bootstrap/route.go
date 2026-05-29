package bootstrap

import (
	authRoutePkg "github.com/KejarBahasa/kejarbill-api/internal/module/auth/route"
	expenseRoutePkg "github.com/KejarBahasa/kejarbill-api/internal/module/expense/route"
	groupRoutePkg "github.com/KejarBahasa/kejarbill-api/internal/module/group/route"
	paymentMethodRoutePkg "github.com/KejarBahasa/kejarbill-api/internal/module/payment_method/route"
	settlementRoutePkg "github.com/KejarBahasa/kejarbill-api/internal/module/settlement/route"
	userRoutePkg "github.com/KejarBahasa/kejarbill-api/internal/module/user/route"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/response"

	"github.com/gofiber/fiber/v3"
)

func RegisterRoute(
	app *fiber.App,
	dep *Dependency,
) {
	app.Get("/", func(c fiber.Ctx) error {
		return response.Success[any](c, "KejarBill API is running", nil)
	})

	v1 := app.Group("/v1")

	// AUTH
	authRoutePkg.AuthRoute(v1, dep.AuthHandler)

	// USER
	userRoutePkg.UserRoute(v1, dep.AuthMiddleware, dep.UserHandler)

	// GROUP
	groupRoutePkg.GroupRoute(
		v1,
		dep.AuthMiddleware,
		dep.GroupHandler,
		dep.GroupMemberHandler,
		dep.GroupParticipantHandler,
		dep.SettlementHandler,
		dep.ActivityHandler,
		dep.BalanceHandler,
		dep.ExpenseHandler,
	)

	// EXPENSE
	expenseRoutePkg.ExpenseRoute(v1, dep.AuthMiddleware, dep.ExpenseHandler)

	// SETTLEMENT
	settlementRoutePkg.SettlementRoute(v1, dep.AuthMiddleware, dep.SettlementHandler)

	// PAYMENT METHOD
	paymentMethodRoutePkg.PaymentMethodRoute(v1, dep.AuthMiddleware, dep.PaymentMethodHandler)
}
