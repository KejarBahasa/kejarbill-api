package route

import (
	"github.com/gofiber/fiber/v3"

	handlerPkg "github.com/KejarBahasa/kejarbill-api/internal/module/payment_method/handler"

	"github.com/KejarBahasa/kejarbill-api/internal/shared/middleware"
)

func PaymentMethodRoute(
	api fiber.Router,
	authMiddleware *middleware.AuthMiddleware,

	handler *handlerPkg.PaymentMethodHandler,
) {
	paymentMethod := api.Group("/payment-methods", authMiddleware.Protected)

	paymentMethod.Post("/", handler.Create)

	paymentMethod.Get("/", handler.FindMine)

	paymentMethod.Patch("/:payment_method_id/default", handler.SetDefault)
	paymentMethod.Patch("/:payment_method_id/hide", handler.Hide)
	paymentMethod.Patch("/:payment_method_id/unhide", handler.Unhide)
}
