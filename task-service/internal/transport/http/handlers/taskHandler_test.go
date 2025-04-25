package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"task-service/internal/middleware"
	"task-service/internal/models"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type MockTaskService struct {
	mock.Mock
}

func (m *MockTaskService) CreateNewTask(ctx context.Context, task models.Task) (int64, error) {
	args := m.Called(ctx, task)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockTaskService) GetTaskByID(ctx context.Context, id int64) (*models.Task, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Task), args.Error(1)
}

func setupTestHandler() (*TaskHandler, *MockTaskService, echo.Context, *httptest.ResponseRecorder) {
	logger := zap.NewNop().Sugar()
	service := new(MockTaskService)
	handler := NewTaskHandler(service, logger)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	ctx := context.WithValue(c.Request().Context(), middleware.LoggerKey, logger)
	c.SetRequest(c.Request().WithContext(ctx))

	return handler, service, c, rec
}

func TestTaskHandler_CreateNewTask_Success(t *testing.T) {
	handler, service, _, _ := setupTestHandler()

	e := echo.New()
	reqBody := models.CreateTaskRequest{
		Type:       "test",
		TemplateID: "template-123",
		Template:   map[string]interface{}{"field": "value"},
		Amount:     5,
		Format:     "json",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", "user-123")

	service.On("CreateNewTask", c.Request().Context(), mock.AnythingOfType("models.Task")).
		Return(int64(1), nil)

	err := handler.CreateNewTask(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var response models.Task
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, int64(1), response.ID)

	service.AssertExpectations(t)
}

func TestTaskHandler_CreateNewTask_InvalidJSON(t *testing.T) {
	handler, _, _, _ := setupTestHandler()

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewReader([]byte("{invalid")))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.CreateNewTask(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response map[string]string
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "invalid request", response["error"])
}

func TestTaskHandler_GetTaskByID_Success(t *testing.T) {
	handler, service, _, _ := setupTestHandler()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/tasks/1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/tasks/:id")
	c.SetParamNames("id")
	c.SetParamValues("1")

	expectedTask := &models.Task{
		ID:        1,
		TaskID:    "task-123",
		UserID:    "user-123",
		Type:      "test",
		Amount:    5,
		Status:    "completed",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	service.On("GetTaskByID", c.Request().Context(), int64(1)).
		Return(expectedTask, nil)

	err := handler.GetTaskByID(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response models.Task
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, expectedTask.ID, response.ID)

	service.AssertExpectations(t)
}
