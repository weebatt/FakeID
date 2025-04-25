package middleware

import (
	"auth-service/pkg/db/redis"
	"auth-service/pkg/logger"
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"golang.org/x/time/rate"
	"net/http"
	"strings"
	"time"
)

// AuthMiddleware verifies JWT tokens in incoming requests
func AuthMiddleware(jwtSecret []byte, redis *redis.Redis, logger *logger.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				logger.Warn("Authorization header missing")
				c.JSON(http.StatusUnauthorized, map[string]interface{}{
					"error": "Authorization header missing",
				})
				return nil
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				logger.Warn("Invalid authorization format")
				c.JSON(http.StatusUnauthorized, map[string]interface{}{
					"error": "Invalid authorization format",
				})
				return nil
			}

			tokenString := parts[1]
			ctx := context.Background()
			_, err := redis.Client.Get(ctx, tokenString).Result()
			if errors.Is(err, redis.Close()) {
				logger.Warn("Token invalidated or expired", "token", tokenString)
				c.JSON(http.StatusUnauthorized, map[string]interface{}{
					"error": "Token invalidated or expired",
				})
				return nil
			} else if err != nil {
				logger.Error("Failed to verify token", "error", err)
				c.JSON(http.StatusInternalServerError, map[string]interface{}{
					"error":   "Failed to verify token",
					"details": err.Error(),
				})
				return nil
			}

			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return jwtSecret, nil
			})

			if err != nil {
				if err == jwt.ErrSignatureInvalid {
					logger.Warn("Invalid token signature", "error", err)
					c.JSON(http.StatusUnauthorized, map[string]interface{}{
						"error":   "Invalid token signature",
						"details": err.Error(),
					})
				} else {
					logger.Warn("Invalid or expired token", "error", err)
					c.JSON(http.StatusUnauthorized, map[string]interface{}{
						"error":   "Invalid or expired token",
						"details": err.Error(),
					})
				}
				return nil
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok || !token.Valid {
				logger.Warn("Invalid token claims")
				c.JSON(http.StatusUnauthorized, map[string]interface{}{
					"error": "Invalid token claims",
				})
				return nil
			}

			if exp, ok := claims["exp"].(float64); ok {
				if time.Now().Unix() > int64(exp) {
					logger.Warn("Token expired")
					c.JSON(http.StatusUnauthorized, map[string]interface{}{
						"error": "Token expired",
					})
					return nil
				}
			}

			logger.Info("User authenticated", "user_id", claims["user_id"], "email", claims["email"])
			c.Set("user_id", claims["user_id"])
			c.Set("email", claims["email"])

			return next(c)
		}
	}
}

// RateLimiter middleware to prevent brute force attacks
func RateLimiter(logger *logger.Logger) echo.MiddlewareFunc {
	limiter := rate.NewLimiter(rate.Every(time.Second), 10)
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if !limiter.Allow() {
				logger.Warn("Rate limit exceeded", "ip", c.RealIP())
				c.JSON(http.StatusTooManyRequests, map[string]interface{}{
					"error": "Too many requests",
				})
				return nil
			}

			return next(c)
		}
	}
}
