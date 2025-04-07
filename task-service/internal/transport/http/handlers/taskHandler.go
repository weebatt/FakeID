package handlers

import (
	"task-service/internal/models"

	"github.com/labstack/echo/v4"
)

type TaskService interface {
	CreateNewTask(task models.Task) error
	GetTaskByID(id int64) error
}

type TaskHandler struct {
	service TaskService
}

func NewTaskHandler(service TaskService) *TaskHandler {
	return &TaskHandler{service: service}
}

func (t *TaskHandler) CreateNewTask(c echo.Context) error {
	var req models.CreateTaskRequest

}

func (t *TaskHandler) GetTaskByID(c echo.Context) error {
	return nil
}
