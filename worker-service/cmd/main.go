package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"worker-service/internal/config"
	http_transport "worker-service/internal/transport/http"

	"worker-service/pkg/logger"
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

	//init router
	routerConfig := http_transport.NewRouterConfig(cfg)
	router := http_transport.NewRouter(routerConfig, log)

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
