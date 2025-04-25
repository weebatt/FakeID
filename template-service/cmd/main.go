package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"template-service/internal/config"
	"template-service/internal/migrations"
	"template-service/internal/repository"
	"template-service/internal/routes"
	"template-service/internal/services"
	http_transport "template-service/internal/transport/http"
	"template-service/internal/transport/http/handlers"
	"template-service/pkg/db/postgres"
	"template-service/pkg/db/redis"
	"template-service/pkg/logger"
	"time"
)

func main() {
	//init config
	cfg, err := config.New()
	if err != nil {
		tempLogger, _ := logger.New("dev")
		tempLogger.Fatal("Failed to initialize config: ", err)
	}

	//init logger
	log, err := logger.New(cfg.Env)
	if err != nil {
		panic(err)
	}
	defer log.Sync()

	//init postgres
	pgClient, err := postgres.NewPostgres(cfg.Postgres, log.SugaredLogger)
	if err != nil {
		log.Fatal("Failed to initialize Postgres: ", err)
	}
	defer pgClient.Close()

	//migrations pgdb
	migrator, err := migrations.New(cfg.Postgres, log.SugaredLogger)
	if err != nil {
		log.Fatalf("Failed to initialize migrator: %w", err)
	}
	if err := migrator.RunMigrations(); err != nil {
		log.Fatalf("Database migration failed: %w", err)
	}

	//init redis
	redisClient, err := redis.NewRedis(cfg.Redis, log.SugaredLogger)
	if err != nil {
		log.Fatal("Failed to initialize Redis: ", err)
	}
	defer redisClient.Close()

	//init router
	routerConfig := http_transport.NewRouterConfig(cfg)
	router := http_transport.NewRouter(routerConfig, log)

	//init repositories
	taskRepository := repository.NewTemplateRepository(pgClient, log.SugaredLogger)

	//init services
	taskService := services.NewTemplateService(taskRepository, redisClient, log.SugaredLogger)

	//init handlers
	taskHandler := handlers.NewTemplateHandler(taskService, log.SugaredLogger)

	//init routes
	routes.SetupTemplateRoutes(router.Echo(), taskHandler)

	//run server
	go func() {
		maxRetries := cfg.HTTPServer.MaxRetries
		retryDelay := time.Duration(cfg.HTTPServer.RetryDelay) * time.Second
		for attempt := 1; attempt <= maxRetries; attempt++ {
			if err := router.Run(); err != nil && err != http.ErrServerClosed {
				log.Errorf("Server failed (attempt %d/%d): retrying in %v...", attempt, maxRetries, retryDelay)
				time.Sleep(retryDelay)
			} else {
				break
			}
		}

		log.Fatalf("Server failed after %d attempts, exiting...", maxRetries)
	}()

	//graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("Received shutdown signal, shutting down gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := router.ShuttingDown(ctx); err != nil {
		log.Errorf("failed to shutdown http server: %s", err)
	}
}
