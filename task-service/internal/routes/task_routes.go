package routes

import (
	"task-service/internal/transport/http/handlers"

	"github.com/labstack/echo/v4"
)

func SetupTaskRoutes(router *echo.Echo, taskHandler *handlers.TaskHandler) {
	api := router.Group("/api/v2/tasks")
	{
		api.POST("", taskHandler.CreateNewTask)
		api.GET("/:id", taskHandler.GetTaskByID)
	}
}
