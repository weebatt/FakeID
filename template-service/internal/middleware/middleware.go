package middleware

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/mock"
)

type ValidateTokenResponse struct {
	UserID string `json:"user_id"`
	Valid  bool   `json:"valid"`
}

type HTTPClient interface {
	Get(url string) (*http.Response, error)
}

type mockHTTPClient struct {
	mock.Mock
}

func (m *mockHTTPClient) Get(url string) (*http.Response, error) {
	args := m.Called(url)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*http.Response), args.Error(1)
}

func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return AuthMiddlewareWithClient(next, http.DefaultClient)
}

func AuthMiddlewareWithClient(next echo.HandlerFunc, client HTTPClient) echo.HandlerFunc {
	return func(c echo.Context) error {
		token := c.Request().Header.Get("Authorization")
		if token == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Missing Authorization header"})
		}

		resp, err := client.Get("http://auth-service:8080/auth/validate?token=" + token)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, map[string]string{"error": "Invalid token"})
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return echo.NewHTTPError(http.StatusUnauthorized, map[string]string{"error": "Invalid token"})
		}

		var validateResp ValidateTokenResponse
		if err := json.NewDecoder(resp.Body).Decode(&validateResp); err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, map[string]string{"error": "Failed to parse token validation response"})
		}

		if !validateResp.Valid {
			return echo.NewHTTPError(http.StatusUnauthorized, map[string]string{"error": "Invalid token"})
		}

		c.Set("user_id", validateResp.UserID)
		return next(c)
	}
}
