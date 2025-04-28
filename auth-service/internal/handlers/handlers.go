package handlers

import (
	"auth-service/internal/models"
	"auth-service/internal/utils"
	postgres "auth-service/pkg/db/postgres"
	"auth-service/pkg/db/redis"
	"auth-service/pkg/logger"
	"context"
	"database/sql"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"net/http"
	"strings"
	"time"
)

type AuthHandler struct {
	db              *postgres.Database
	redis           *redis.Redis
	jwtSecret       []byte
	tokenExpiration time.Duration
	logger          *logger.Logger
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(db *postgres.Database, redis *redis.Redis, jwtSecret []byte, tokenExpiration time.Duration, logger *logger.Logger) *AuthHandler {
	return &AuthHandler{
		db:              db,
		redis:           redis,
		jwtSecret:       jwtSecret,
		tokenExpiration: tokenExpiration,
		logger:          logger,
	}
}

// Register handles user registration
func (h *AuthHandler) Register(c echo.Context) error {
	var user models.UserRegister

	// Validate input JSON
	if err := c.Bind(&user); err != nil {
		h.logger.Error("Invalid input format", "error", err)
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "Invalid input format",
			"details": err.Error(),
		})
		return nil
	}

	// Additional validation
	if err := user.Validate(); err != nil {
		h.logger.Warn("Invalid additional validation", "email", user.Email, "error", err)
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "Invalid additional validation",
			"details": err.Error(),
		})
		return nil
	}

	// Check if user already exists
	var exists bool
	err := h.db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)",
		user.Email).Scan(&exists)
	if err != nil {
		h.logger.Error("Database error checking user existence", "email", user.Email, "error", err)
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "Database error",
			"details": err.Error(),
		})
		return nil
	}
	if exists {
		h.logger.Warn("Email already registered", "email", user.Email)
		c.JSON(http.StatusConflict, map[string]interface{}{
			"error": "Email already registered",
		})
		return nil
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		h.logger.Error("Password hashing failed", "error", err)
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "Password processing failed",
			"details": err.Error(),
		})
		return nil
	}

	// Insert user with transaction
	tx, err := h.db.DB.Begin()
	if err != nil {
		h.logger.Error("Transaction start failed", "error", err)
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "Transaction start failed",
			"details": err.Error(),
		})
		return nil
	}

	var id int
	err = tx.QueryRow(`
        INSERT INTO users (email, password_hash) 
        VALUES ($1, $2) 
        RETURNING id`,
		user.Email, hashedPassword,
	).Scan(&id)

	if err != nil {
		tx.Rollback()
		h.logger.Error("User creation failed", "email", user.Email, "error", err)
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "User creation failed",
			"details": err.Error(),
		})
		return nil
	}

	if err = tx.Commit(); err != nil {
		h.logger.Error("Transaction commit failed", "error", err)
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "Transaction commit failed",
			"details": err.Error(),
		})
		return nil
	}

	h.logger.Info("User registered successfully", "user_id", id, "email", user.Email)
	c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "User registered successfully",
		"user_id": id,
	})
	return nil
}

// Login handles user authentication and JWT generation
func (h *AuthHandler) Login(c echo.Context) error {
	var login models.UserLogin
	if err := c.Bind(&login); err != nil {
		h.logger.Error("Invalid login data", "error", err)
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "Invalid login data",
			"details": err.Error(),
		})
		return nil
	}

	// Get user from database
	var user models.User
	err := h.db.DB.QueryRow(`
        SELECT id, email, password_hash 
        FROM users 
        WHERE email = $1`,
		login.Email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash)

	if err == sql.ErrNoRows {
		h.logger.Warn("Invalid login attempt", "email", login.Email)
		c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"error": "Invalid credentials",
		})
		return nil
	}
	if err != nil {
		h.logger.Error("Login process failed", "email", login.Email, "error", err)
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "Login process failed",
			"details": err.Error(),
		})
		return nil
	}

	// Verify password
	if !utils.CheckPasswordHash(login.Password, user.PasswordHash) {
		h.logger.Warn("Invalid password attempt", "email", login.Email)
		c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"error": "Invalid credentials",
		})
		return nil
	}

	// Generate JWT with claims
	now := time.Now()
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"iat":     now.Unix(),
		"exp":     now.Add(h.tokenExpiration).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(h.jwtSecret)
	if err != nil {
		h.logger.Error("Token generation failed", "user_id", user.ID, "error", err)
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "Token generation failed",
			"details": err.Error(),
		})
		return nil
	}

	// Save token in Redis with expiration
	ctx := context.Background()
	err = h.redis.Client.Set(ctx, tokenString, user.ID, h.tokenExpiration).Err()
	if err != nil {
		h.logger.Error("Failed to save token in Redis", "user_id", user.ID, "error", err)
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "Failed to save token",
			"details": err.Error(),
		})
		return nil
	}

	h.logger.Info("User logged in successfully", "user_id", user.ID, "email", user.Email)
	c.JSON(http.StatusOK, map[string]interface{}{
		"token":      tokenString,
		"expires_in": h.tokenExpiration.Seconds(),
		"token_type": "Bearer",
	})
	return nil
}

// RefreshToken generates a new token for valid users
func (h *AuthHandler) RefreshToken(c echo.Context) error {
	userID, ok := c.Get("user_id").(float64)
	if !ok {
		h.logger.Warn("User not authenticated")
		c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"error": "User not authenticated",
		})
		return nil
	}

	email, ok := c.Get("email").(string)
	if !ok {
		h.logger.Warn("Invalid user data")
		c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"error": "Invalid user data",
		})
		return nil
	}

	now := time.Now()
	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"iat":     now.Unix(),
		"exp":     now.Add(h.tokenExpiration).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(h.jwtSecret)
	if err != nil {
		h.logger.Error("Token refresh failed", "user_id", userID, "error", err)
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "Token refresh failed",
			"details": err.Error(),
		})
		return nil
	}

	ctx := context.Background()
	err = h.redis.Client.Set(ctx, tokenString, userID, h.tokenExpiration).Err()
	if err != nil {
		h.logger.Error("Failed to save refreshed token", "user_id", userID, "error", err)
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "Failed to save refreshed token",
			"details": err.Error(),
		})
		return nil
	}

	h.logger.Info("Token refreshed successfully", "user_id", userID, "email", email)
	c.JSON(http.StatusOK, map[string]interface{}{
		"token":      tokenString,
		"expires_in": h.tokenExpiration.Seconds(),
		"token_type": "Bearer",
	})
	return nil
}

// Logout invalidates the token
func (h *AuthHandler) Logout(c echo.Context) error {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		h.logger.Warn("Authorization header missing")
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Authorization header missing",
		})
		return nil
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		h.logger.Warn("Invalid authorization format")
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid authorization format",
		})
		return nil
	}

	tokenString := parts[1]
	ctx := context.Background()
	err := h.redis.Client.Del(ctx, tokenString).Err()
	if err != nil {
		h.logger.Error("Failed to invalidate token", "error", err)
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "Failed to invalidate token",
			"details": err.Error(),
		})
		return nil
	}

	h.logger.Info("User logged out successfully")
	c.JSON(http.StatusOK, map[string]interface{}{
		"message":      "Successfully logged out",
		"instructions": "Please remove the token from your client storage",
	})
	return nil
}
