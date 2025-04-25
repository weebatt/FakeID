package routes

import (
	"template-service/internal/middleware"
	"template-service/internal/transport/http/handlers"

	"github.com/labstack/echo"
)

func SetupTemplateRoutes(router *echo.Echo, templateHandler *handlers.TemplateHandler) {
	group := router.Group("/templates", middleware.AuthMiddleware)
	{
		group.POST("", templateHandler.CreateNewTemplate)
		group.GET("/:id", templateHandler.GetTemplateByID)
	}
}
