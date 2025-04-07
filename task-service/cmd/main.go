package main

import (
	"task-service/internal/config"
	"task-service/internal/transport/http"
)

func main() {
	//init config
	cfg, err := config.New()
	if err != nil {
		panic(err)
	}

	//inti logger
	logger := logger.NewDevLogger()

	//init mongo

	//init router
	routerConfig := http.NewRouterConfig(cfg)
	router := http.NewRouter(routerConfig)

	//run router
	if err := router.Run(); err != nil {
		panic(err)
	}
}
