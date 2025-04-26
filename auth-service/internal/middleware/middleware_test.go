package middleware

//import (
//	"net/http"
//	"net/http/httptest"
//	"testing"
//
//	"github.com/labstack/echo/v4"
//)
//
//func TestRateLimiter(t *testing.T) {
//	e := echo.New()
//	h := RateLimiter(func(c echo.Context) error { return c.String(http.StatusOK, "ok") })
//
//	req := httptest.NewRequest(http.MethodGet, "/", nil)
//	rec := httptest.NewRecorder()
//	c := e.NewContext(req, rec)
//
//	if err := h(c); err != nil {
//		t.Fatalf("first call failed: %v", err)
//	}
//	if rec.Code != http.StatusOK {
//		t.Fatalf("want 200, got %d", rec.Code)
//	}
//}
