package middleware

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware(t *testing.T) {
	e := echo.New()

	tests := []struct {
		name           string
		token          string
		mockResponse   *http.Response
		mockError      error
		expectCall     bool
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "missing token",
			token:          "",
			expectCall:     false,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"Missing Authorization header"}`,
		},
		{
			name:           "auth service error",
			token:          "valid.token.here",
			mockError:      errors.New("service unavailable"),
			expectCall:     true,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"Invalid token"}`,
		},
		{
			name:           "invalid token response",
			token:          "invalid.token.here",
			mockResponse:   &http.Response{StatusCode: http.StatusUnauthorized},
			expectCall:     true,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"Invalid token"}`,
		},
		{
			name:  "invalid json response",
			token: "bad.json.token",
			mockResponse: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString("invalid json")),
			},
			expectCall:     true,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"Failed to parse token validation response"}`,
		},
		{
			name:  "invalid token",
			token: "invalid.token",
			mockResponse: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(`{"user_id":"123","valid":false}`)),
			},
			expectCall:     true,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"Invalid token"}`,
		},
		{
			name:  "valid token",
			token: "valid.token",
			mockResponse: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(`{"user_id":"123","valid":true}`)),
			},
			expectCall:     true,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(mockHTTPClient)

			if tt.expectCall {
				mockClient.On("Get", "http://auth-service:8080/auth/validate?token="+tt.token).
					Return(tt.mockResponse, tt.mockError)
			}

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", tt.token)
			}
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			testHandler := func(c echo.Context) error {
				if tt.expectedStatus == http.StatusOK {
					userID := c.Get("user_id").(string)
					assert.Equal(t, "123", userID)
				}
				return c.String(http.StatusOK, "OK")
			}

			handler := AuthMiddlewareWithClient(testHandler, mockClient)
			err := handler(c)

			if tt.expectedStatus != http.StatusOK {
				if assert.Error(t, err) {
					he, ok := err.(*echo.HTTPError)
					if assert.True(t, ok, "Expected HTTPError") {
						assert.Equal(t, tt.expectedStatus, he.Code)
						if body, ok := he.Message.(map[string]string); ok {
							assert.JSONEq(t, tt.expectedBody, `{"error":"`+body["error"]+`"}`)
						} else {
							t.Errorf("Unexpected message type: %T", he.Message)
						}
					}
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, http.StatusOK, rec.Code)
			}

			if tt.expectCall {
				mockClient.AssertExpectations(t)
			}
		})
	}
}
