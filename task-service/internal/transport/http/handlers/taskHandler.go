package handlers

import (
	"context"
	"net/http"
	"strconv"
	"task-service/internal/middleware"
	"task-service/internal/models"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type TaskService interface {
	CreateNewTask(ctx context.Context, task models.Task) (int64, error)
	GetTaskByID(ctx context.Context, id int64) (*models.Task, error)
	ListTasks(ctx context.Context, filter models.TaskFilter) ([]models.Task, error)
}

type TaskHandler struct {
	service TaskService
	logger  *zap.SugaredLogger
}

func NewTaskHandler(service TaskService, logger *zap.SugaredLogger) *TaskHandler {
	return &TaskHandler{service: service, logger: logger}
}

func (t *TaskHandler) CreateNewTask(c echo.Context) error {
	var req models.CreateTaskRequest

	ctxLogger := middleware.GetLoggerFromCtx(c.Request().Context())

	if err := c.Bind(&req); err != nil {
		ctxLogger.Errorf("Invalid request: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		ctxLogger.Errorf("Validation failed: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	task := models.Task{
		TaskID:     uuid.NewString(), // Генерируем UUID для TaskID
		UserID:     c.Get("user_id").(string),
		Type:       req.Type,
		TemplateID: req.TemplateID,
		Template:   req.Template,
		Amount:     req.Amount,
	}

	id, err := t.service.CreateNewTask(c.Request().Context(), task)
	if err != nil {
		ctxLogger.Errorf("Failed to create task: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create task"})
	}

	task.ID = id
	ctxLogger.Infof("Task created with ID: %d, TaskID: %s", id, task.TaskID)
	return c.JSON(http.StatusCreated, task)
}

func (t *TaskHandler) GetTaskByID(c echo.Context) error {
	id := c.Param("id")
	intID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid task ID"})
	}

	logger := middleware.GetLoggerFromCtx(c.Request().Context())

	if intID < 0 {
		logger.Errorf("Invalid task ID")
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid task ID"})
	}

	task, err := t.service.GetTaskByID(c.Request().Context(), intID)
	if err != nil {
		logger.Errorf("Failed to get task %s: %v", id, err)
		return c.JSON(http.StatusNotFound, map[string]string{"error": "task not found"})
	}

	return c.JSON(http.StatusOK, task)
}

func (t *TaskHandler) ListTasks(c echo.Context) error {
	ctx := c.Request().Context()
	logger := middleware.GetLoggerFromCtx(ctx)

	// Get query parameters for filtering
	userID := c.QueryParam("user_id")
	taskType := c.QueryParam("type")
	status := c.QueryParam("status")

	// Extract pagination parameters
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page <= 0 {
		page = 1
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 || limit > 100 {
		limit = 10 // Default limit
	}

	// Log the request details
	logger.Infow("Listing tasks",
		"user_id", userID,
		"type", taskType,
		"status", status,
		"page", page,
		"limit", limit,
	)

	// Retrieve tasks from the service
	tasks, err := t.service.ListTasks(ctx, models.TaskFilter{
		UserID: userID,
		Type:   taskType,
		Status: status,
		Page:   page,
		Limit:  limit,
	})

	if err != nil {
		logger.Errorw("Failed to list tasks", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to retrieve tasks"})
	}

	// Return the list of tasks
	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":  tasks,
		"page":  page,
		"limit": limit,
	})
}
