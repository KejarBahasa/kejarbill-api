package middleware

import (
	"context"
	"time"

	"github.com/KejarBahasa/kejarbill-api/internal/shared/constants"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/logger"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/utils"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog"
)

func ZeroLog() fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := c.Context()
		start := time.Now()

		err := c.Next()

		latency := time.Since(start)

		requestID, _ := ctx.Value(constants.CtxKeyRequestID).(string)

		clientInfo := utils.GetClientInfo(c)

		logEvent := logger.Log.Info()

		if err != nil {
			logEvent = logger.Log.Error().Err(err)
		}

		logEvent.
			Str("request_id", requestID).
			Str("method", c.Method()).
			Str("path", c.OriginalURL()).
			Int("status", c.Response().StatusCode()).
			Dur("latency", latency).
			Str("ip", clientInfo.IPAddress).
			Str("user_agent", clientInfo.UserAgent).
			Str("client_type", clientInfo.ClientType).
			Msg("http request")

		return err
	}
}

func LoggerFromContext(ctx context.Context) *zerolog.Logger {
	log, ok := ctx.Value(constants.CtxKeyLogger).(*zerolog.Logger)
	if !ok {
		defaultLogger := logger.Log
		return &defaultLogger
	}

	return log
}
