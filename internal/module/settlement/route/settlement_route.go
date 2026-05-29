package route

import (
	"github.com/KejarBahasa/kejarbill-api/internal/module/settlement/handler"

	"github.com/KejarBahasa/kejarbill-api/internal/shared/middleware"

	"github.com/gofiber/fiber/v3"
)

func SettlementRoute(
	api fiber.Router,

	authMiddleware *middleware.AuthMiddleware,

	handler *handler.SettlementHandler,
) {
	settlement := api.Group("/settlement", authMiddleware.Protected)

	settlement.Get("/:settlement_id", handler.GetDetail)
}
