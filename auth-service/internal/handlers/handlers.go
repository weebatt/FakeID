package handlers

import (
	"context"
	"database/sql"
	database "fake_id/internal/db"
	"fake_id/internal/models"
	"fake_id/internal/redis"
	"fake_id/internal/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"net/http"
	"strings"
	"time"
)

type AuthHandler struct {
	db              *database.Database
	redis           *redis.Redis
	jwtSecret       []byte
	tokenExpiration time.Duration
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(db *database.Database, redis *redis.Redis, jwtSecret []byte, tokenExpiration time.Duration) *AuthHandler {
	return &AuthHandler{
		db:              db,
		redis:           redis,
		jwtSecret:       jwtSecret,
		tokenExpiration: tokenExpiration,
	}
}

// Register handles user registration
func (h *AuthHandler) Register(c echo.Context) error {
	var user models.UserRegister

	// Validate input JSON
	if err := c.Bind(&user); err != nil {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "Invalid input format",
			"details": err.Error(),
		})
		return nil
	}

	// Additional validation
	if err := user.Validate(); err != nil {
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
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "Database error",
			"details": err.Error(),
		})
		return nil
	}
	if exists {
		c.JSON(http.StatusConflict, map[string]interface{}{
			"error": "Email already registered",
		})
		return nil
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "Password processing failed",
			"details": err.Error(),
		})
		return nil
	}

	// Insert user with transaction
	tx, err := h.db.DB.Begin()
	if err != nil {
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
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "User creation failed",
			"details": err.Error(),
		})
		return nil
	}

	if err = tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "Transaction commit failed",
			"details": err.Error(),
		})
		return nil
	}

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
		c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"error": "Invalid credentials",
		})
		return nil
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "Login process failed",
			"details": err.Error(),
		})
		return nil
	}

	// Verify password
	if !utils.CheckPasswordHash(login.Password, user.PasswordHash) {
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
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "Failed to save token",
			"details": err.Error(),
		})
		return nil
	}

	// Return token with expiration
	c.JSON(http.StatusOK, map[string]interface{}{
		"token":      tokenString,
		"expires_in": h.tokenExpiration.Seconds(),
		"token_type": "Bearer",
	})
	return nil
}

// RefreshToken generates a new token for valid users
func (h *AuthHandler) RefreshToken(c echo.Context) error {
	// Get user ID from context (set by auth middleware)
	userID, ok := c.Get("user_id").(float64)
	if !ok {
		c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"error": "User not authenticated",
		})
		return nil
	}

	// Get email from context
	email, ok := c.Get("email").(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"error": "Invalid user data",
		})
		return nil
	}

	// Generate new token
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
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "Token refresh failed",
			"details": err.Error(),
		})
		return nil
	}

	// Save new token in Redis
	ctx := context.Background()
	err = h.redis.Client.Set(ctx, tokenString, userID, h.tokenExpiration).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "Failed to save refreshed token",
			"details": err.Error(),
		})
		return nil
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"token":      tokenString,
		"expires_in": h.tokenExpiration.Seconds(),
		"token_type": "Bearer",
	})
	return nil
}

// Logout invalidates the token
func (h *AuthHandler) Logout(c echo.Context) error {
	// Get token from Authorization header
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Authorization header missing",
		})
		return nil
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid authorization format",
		})
		return nil
	}

	tokenString := parts[1]

	// Invalidate token by removing it from Redis
	ctx := context.Background()
	err := h.redis.Client.Del(ctx, tokenString).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "Failed to invalidate token",
			"details": err.Error(),
		})
		return nil
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"message":      "Successfully logged out",
		"instructions": "Please remove the token from your client storage",
	})
	return nil
}
