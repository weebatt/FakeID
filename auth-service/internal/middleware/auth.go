package middleware

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	_ "github.com/labstack/echo/v4"
	"golang.org/x/time/rate"
	"net/http"
	"strings"
	"time"
)

// AuthMiddleware verifies JWT tokens in incoming requests
func AuthMiddleware(jwtSecret []byte) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get Authorization header
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				c.JSON(http.StatusUnauthorized, map[string]interface{}{
					"error": "Authorization header missing",
				})
				return nil
			}

			// Check Bearer scheme
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				c.JSON(http.StatusUnauthorized, map[string]interface{}{
					"error": "Invalid authorization format",
				})
				return nil
			}

			tokenString := parts[1]

			// Parse and validate token
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				// Validate signing method
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return jwtSecret, nil
			})

			if err != nil {
				if err == jwt.ErrSignatureInvalid {
					c.JSON(http.StatusUnauthorized, map[string]interface{}{
						"error":   "Invalid token signature",
						"details": err.Error(),
					})
				} else {
					c.JSON(http.StatusUnauthorized, map[string]interface{}{
						"error":   "Invalid or expired token",
						"details": err.Error(),
					})
				}
				return nil
			}

			// Extract and validate claims
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok || !token.Valid {
				c.JSON(http.StatusUnauthorized, map[string]interface{}{
					"error":   "Invalid token claims",
					"details": err.Error(),
				})
				return nil
			}

			// Check token expiration
			if exp, ok := claims["exp"].(float64); ok {
				if time.Now().Unix() > int64(exp) {
					c.JSON(http.StatusUnauthorized, map[string]interface{}{
						"error":   "Token expired",
						"details": err.Error(),
					})
					return nil
				}
			}

			// Set user information in context
			c.Set("user_id", claims["user_id"])
			c.Set("email", claims["email"])

			if err := next(c); err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{
					"error": err.Error(),
				})
			}
			return nil
		}
	}
}

// RateLimiter middleware to prevent brute force attacks
func RateLimiter(next echo.HandlerFunc) echo.HandlerFunc {
	limiter := rate.NewLimiter(rate.Every(time.Second), 10)
	return func(c echo.Context) error {
		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, map[string]interface{}{
				"error": "Too many requests",
			})
			return nil
		}

		if err := next(c); err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{
				"error": err.Error(),
			})
		}
		return nil
	}
}
