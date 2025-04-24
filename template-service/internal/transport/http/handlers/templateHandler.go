package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"template-service/internal/models"
	"template-service/internal/services"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo"
	"go.uber.org/zap"
)

type TemplateService interface {
	CreateNewTemplate(ctx context.Context, template models.Template) (int64, error)
	GetTemplateByID(ctx context.Context, id int64) (*models.Template, error)
}

type TemplateHandler struct {
	service services.TemplateService
	logger  *zap.SugaredLogger
}

func NewTemplateHandler(service services.TemplateService, logger *zap.SugaredLogger) *TemplateHandler {
	return &TemplateHandler{service: service, logger: logger}
}

func (h *TemplateHandler) CreateNewTemplate(c echo.Context) error {
	var req models.CreateTemplateRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		h.logger.Errorf("Failed to decode request body: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		h.logger.Errorf("Validation failed: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	template := models.Template{
		Title:   req.Title,
		Content: req.Content,
	}

	id, err := h.service.CreateNewTemplate(c.Request().Context(), template)
	if err != nil {
		h.logger.Errorf("Failed to create template: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create template"})
	}

	return c.JSON(http.StatusCreated, map[string]string{"id": strconv.FormatInt(id, 10)})
}

func (h *TemplateHandler) GetTemplateByID(c echo.Context) error {
	id := c.Param("id")
	intID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid template ID"})
	}

	template, err := h.service.GetTemplateByID(c.Request().Context(), intID)
	if err != nil {
		h.logger.Errorf("Failed to get template %s: %v", id, err)
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Template not found"})
	}

	return c.JSON(http.StatusOK, template)
}
