package middleware

import (
	"context"
	"task-service/pkg/logger"
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
			fields := []interface{}{
				"status", c.Response().Status,
				"latency", stop.String(),
			}
			if err != nil {
				fields = append(fields, "error", err.Error())
				logger.Errorw("Request failed", fields...)
			} else {
				logger.Infow("Request completed", fields...)
			}
			return err
		}
	}
}

func GetLoggerFromCtx(ctx context.Context) *zap.SugaredLogger {
	log, ok := ctx.Value(LoggerKey).(*zap.SugaredLogger)
	if !ok {
		l, err := logger.New("prod")
		if err != nil {
			l, _ := zap.NewProduction()
			log = l.Sugar()
			log.Warn("Failed to create fallback logger, using minimal logger")
		} else {
			log = l.SugaredLogger
			log.Warn("Logger not found in context, using fallback prod logger")
		}
	}
	return log
}

func GetRequestIDFromCtx(ctx context.Context) string {
	requestID, ok := ctx.Value(RequestIDKey).(string)
	if !ok {
		return ""
	}

	return requestID
}
