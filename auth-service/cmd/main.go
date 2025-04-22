package main

import (
	"fake_id/internal/config"
	database "fake_id/internal/db"
	"fake_id/internal/handlers"
	"fake_id/internal/middleware"
	"fake_id/internal/redis"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"net/http"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Initialize database
	db, err := database.NewDatabase(cfg.GetDSN())
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.DB.Close()

	// Initialize Redis
	redisClient, err := redis.NewRedis(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}
	defer redisClient.Close()

	r := echo.New()
	if cfg.Environment == "production" {
		// отключить debug-режим
		r.Debug = false
		// скрыть баннер и сообщение о порте
		r.HideBanner = true
		r.HidePort = true
		// оставить только ERROR‑логи
		if l, ok := r.Logger.(*log.Logger); ok {
			l.SetLevel(log.ERROR)
		}
	}

	// CORS middleware
	r.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set("Access-Control-Allow-Origin", "*")
			c.Response().Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			c.Response().Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if c.Request().Method == "OPTIONS" {
				c.Response().Writer.WriteHeader(http.StatusNoContent)
				return c.NoContent(http.StatusNoContent)
			}
			return next(c)
		}
	})

	// Initialize handlers with JWT configuration and Redis
	authHandler := handlers.NewAuthHandler(db, redisClient, []byte(cfg.JWT.Secret), cfg.JWT.TokenExpiry)

	// Public routes
	public := r.Group("/api/v1")
	{
		public.POST("/register", authHandler.Register)
		public.POST("/login", authHandler.Login)
	}

	// Protected routes with JWT middleware
	protected := r.Group("/api/v1")
	protected.Use(middleware.AuthMiddleware([]byte(cfg.JWT.Secret), redisClient))
	{
		protected.POST("/refresh-token", authHandler.RefreshToken)
		protected.POST("/logout", authHandler.Logout)
		protected.GET("/profile", getUserProfile)
	}

	// Start server with configured host and port
	serverAddr := cfg.Server.Host + ":" + cfg.Server.Port
	log.Printf("Server starting on %s", serverAddr)

	srv := &http.Server{
		Addr:         serverAddr,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("Server failed to start:", err)
	}
}

func getUserProfile(c echo.Context) error {
	userID := c.Get("user_id")
	email := c.Get("email")

	c.JSON(200, map[string]interface{}{
		"user_id": userID,
		"email":   email,
	})
	return nil
}
