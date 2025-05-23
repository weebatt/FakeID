package http

import (
	"context"
	"fmt"
	"task-service/internal/config"
	"task-service/internal/middleware"
	"task-service/pkg/logger"

	"github.com/labstack/echo/v4"
)

type RouterConfig struct {
	Host string
	Port string
}

type Router struct {
	config RouterConfig
	router *echo.Echo
}

func NewRouterConfig(cfg *config.Config) RouterConfig {
	return RouterConfig{
		Host: cfg.HTTPServer.Host,
		Port: cfg.HTTPServer.Port,
	}
}

func NewRouter(rConfig RouterConfig, log *logger.Logger) *Router {
	r := echo.New()
	r.Use(middleware.LoggerMiddleware(log.SugaredLogger))
	r.Use(middleware.RequestLogger())
	return &Router{
		config: rConfig,
		router: r,
	}
}

func (r *Router) Run() error {
	return r.router.Start(fmt.Sprintf("%s:%s", r.config.Host, r.config.Port))
}

func (r *Router) ShuttingDown(ctx context.Context) error {
	return r.router.Shutdown(ctx)
}

func (r *Router) Echo() *echo.Echo {
	return r.router
}
