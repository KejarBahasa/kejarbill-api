package route

import (
	"github.com/KejarBahasa/kejarbill-api/internal/module/group/handler"
	balanceHandlerPkg "github.com/KejarBahasa/kejarbill-api/internal/module/ledger/handler"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/middleware"

	"github.com/gofiber/fiber/v3"
)

func GroupRoute(
	api fiber.Router,
	authMiddleware *middleware.AuthMiddleware,
	groupHandler *handler.GroupHandler,
	balanceHandler *balanceHandlerPkg.BalanceHandler,
) {
	group := api.Group("/groups", authMiddleware.Protected)

	group.Post("/", groupHandler.Create)

	group.Get("/:group_id/balances", balanceHandler.GetGroupBalances)
}
