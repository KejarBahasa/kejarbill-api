package middleware

import (
	"context"

	"github.com/KejarBahasa/kejarbill-api/internal/shared/constants"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/logger"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/requestid"
)

func ContextBridge() fiber.Handler {
	return func(c fiber.Ctx) error {
		requestID := requestid.FromContext(c)

		ctx := context.WithValue(
			c.Context(),
			constants.CtxKeyRequestID,
			requestID,
		)

		reqLogger := logger.Log.With().
			Str("request_id", requestID).
			Logger()

		ctx = context.WithValue(
			ctx,
			constants.CtxKeyLogger,
			&reqLogger,
		)

		c.SetContext(ctx)

		return c.Next()
	}
}
