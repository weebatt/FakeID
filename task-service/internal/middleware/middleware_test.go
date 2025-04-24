package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func newTestLogger() *zap.SugaredLogger {
	cfg := zap.NewDevelopmentConfig()
	cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	logger, _ := cfg.Build()
	return logger.Sugar()
}

func TestLoggerMiddleware_SetsRequestIDAndLogger(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	m := LoggerMiddleware(newTestLogger())
	handler := m(func(c echo.Context) error {
		reqID := GetRequestIDFromCtx(c.Request().Context())
		log := GetLoggerFromCtx(c.Request().Context())

		assert.NotEmpty(t, reqID)
		assert.NotNil(t, log)

		return c.String(http.StatusOK, "ok")
	})

	err := handler(ctx)
	assert.NoError(t, err)

	assert.NotEmpty(t, rec.Header().Get("X-Request-ID"))
}

func TestRequestLogger_LogsSuccessAndError(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	log := newTestLogger()
	ctx.SetRequest(req.WithContext(
		context.WithValue(req.Context(), LoggerKey, log),
	))

	m := RequestLogger()

	t.Run("success", func(t *testing.T) {
		h := m(func(c echo.Context) error {
			return c.String(http.StatusOK, "ok")
		})
		err := h(ctx)
		assert.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		h := m(func(c echo.Context) error {
			return echo.NewHTTPError(http.StatusBadRequest, "bad request")
		})
		err := h(ctx)
		assert.Error(t, err)
	})
}
