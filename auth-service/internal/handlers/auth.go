package handlers

import (
	"database/sql"
	database "fake_id/internal/db"
	"fake_id/internal/models"
	"fake_id/internal/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	_ "github.com/labstack/echo/v4"
	"net/http"
	"time"
)

type AuthHandler struct {
	db        *database.Database
	jwtSecret []byte
	// Add token expiration configuration
	tokenExpiration time.Duration
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(db *database.Database, jwtSecret []byte) *AuthHandler {
	return &AuthHandler{
		db:              db,
		jwtSecret:       jwtSecret,
		tokenExpiration: 24 * time.Hour, // Default 24 hour expiration
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
			"error":   "Email already registered",
			"details": err.Error(),
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
		// Don't specify whether email or password was wrong
		c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"error":   "Invalid credentials",
			"details": err.Error(),
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
		// Use same message as above for security
		c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"error":   "Invalid credentials",
			"details": err.Error(),
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
	userID := c.Get("user_id")
	if userID == nil {
		c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"error": "User not authenticated",
		})
		return nil
	}

	// Generate new token
	now := time.Now()
	claims := jwt.MapClaims{
		"user_id": userID,
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

	c.JSON(http.StatusOK, map[string]interface{}{
		"token":      tokenString,
		"expires_in": h.tokenExpiration.Seconds(),
		"token_type": "Bearer",
	})
	return nil
}

// Logout endpoint (optional - useful for client-side cleanup)
func (h *AuthHandler) Logout(c echo.Context) error {
	// Since JWT is stateless, server-side logout isn't needed
	// However, we can return instructions for the client
	c.JSON(http.StatusOK, map[string]interface{}{
		"message":      "Successfully logged out",
		"instructions": "Please remove the token from your client storage",
	})
	return nil
}
