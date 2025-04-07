package middleware

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type ctxKey string

const LoggerKey ctxKey = "logger"
const RequestIDKey ctxKey = "request_id"

func LoggerMiddleware(logger *zap.SugaredLogger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()

			//generate request id
			requestID := c.Request().Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = uuid.NewString()
			}

			//create context
			ctx := context.WithValue(c.Request().Context(), RequestIDKey, requestID)

			//add logger
			enrichedLogger := logger.With(
				"request_id", requestID,
				"method", req.Method,
				"url", req.URL.String(),
				"remote", c.RealIP(),
			)
			ctx = context.WithValue(ctx, LoggerKey, enrichedLogger)

			c.SetRequest(req.WithContext(ctx))
			c.Response().Header().Set("X-Request-ID", requestID)

			return next(c)
		}
	}
}

func RequestLogger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)
			stop := time.Since(start)

			ctx := c.Request().Context()
			logger := GetLoggerFromCtx(ctx)

			logger.Infow("request completed",
				"status", c.Response().Status,
				"latency", stop.String(),
			)

			return err
		}
	}
}

func GetLoggerFromCtx(ctx context.Context) *zap.SugaredLogger {
	logger, ok := ctx.Value(LoggerKey).(*zap.SugaredLogger)
	if !ok {
		return zap.NewNop().Sugar()
	}

	return logger
}

func GetRequestIDFromCtx(ctx context.Context) string {
	requestID, ok := ctx.Value(RequestIDKey).(string)
	if !ok {
		return ""
	}

	return requestID
}
