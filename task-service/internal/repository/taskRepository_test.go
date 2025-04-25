package repository

import (
	"context"
	"errors"
	"task-service/internal/models"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// fakeDB — фейковая реализация интерфейса DB для тестов.
type fakeDB struct {
	mock pgxmock.PgxPoolIface
}

func (f *fakeDB) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return f.mock.QueryRow(ctx, sql, args...)
}

func (f *fakeDB) Exec(ctx context.Context, sql string, args ...interface{}) error {
	_, err := f.mock.Exec(ctx, sql, args...)
	return err
}

func (f *fakeDB) Ping(ctx context.Context) error {
	return f.mock.Ping(ctx)
}

func (f *fakeDB) Close() {
	f.mock.Close()
}

func setupTaskRepository(t *testing.T) (*postgresTaskRepository, pgxmock.PgxPoolIface) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)

	logger, _ := zap.NewDevelopment()
	sugaredLogger := logger.Sugar()

	// Создаем фейковую реализацию DB
	db := &fakeDB{mock: mock}
	repo := NewTaskRepository(db, sugaredLogger)
	return repo, mock
}

// TestCreateNewTask_Success проверяет успешное создание задачи.
func TestCreateNewTask_Success(t *testing.T) {
	repo, mock := setupTaskRepository(t)
	defer mock.Close()

	task := models.Task{
		UserID:     "user-123",
		Type:       "test",
		TemplateID: "template-456",
		Template:   map[string]interface{}{"name": "{{name}}"},
		Amount:     100,
	}

	mock.ExpectQuery(`INSERT INTO tasks`).
		WithArgs(pgxmock.AnyArg(), "user-123", "test", "template-456", task.Template, 100, "pending", pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(int64(1)))

	id, err := repo.CreateNewTask(context.Background(), task)
	require.NoError(t, err)
	assert.Equal(t, int64(1), id)

	require.NoError(t, mock.ExpectationsWereMet())
}

// TestCreateNewTask_DBError проверяет ошибку базы данных при создании задачи.
func TestCreateNewTask_DBError(t *testing.T) {
	repo, mock := setupTaskRepository(t)
	defer mock.Close()

	task := models.Task{
		UserID:     "user-123",
		Type:       "test",
		TemplateID: "template-456",
		Template:   map[string]interface{}{"name": "{{name}}"},
		Amount:     100,
	}

	mock.ExpectQuery(`INSERT INTO tasks`).
		WithArgs(pgxmock.AnyArg(), "user-123", "test", "template-456", task.Template, 100, "pending", pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnError(errors.New("db error"))

	id, err := repo.CreateNewTask(context.Background(), task)
	require.Error(t, err)
	assert.Equal(t, int64(0), id)
	assert.Contains(t, err.Error(), "db error")

	require.NoError(t, mock.ExpectationsWereMet())
}

// TestGetTaskByID_Success проверяет успешное получение задачи.
func TestGetTaskByID_Success(t *testing.T) {
	repo, mock := setupTaskRepository(t)
	defer mock.Close()

	task := models.Task{
		ID:         1,
		TaskID:     "task-123",
		UserID:     "user-123",
		Type:       "test",
		TemplateID: "template-456",
		Template:   map[string]interface{}{"name": "{{name}}"},
		Amount:     100,
		Status:     "pending",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	mock.ExpectQuery(`SELECT id, task_id, user_id, type, template_id, template, amount, status, created_at, updated_at`).
		WithArgs(int64(1)).
		WillReturnRows(pgxmock.NewRows([]string{"id", "task_id", "user_id", "type", "template_id", "template", "amount", "status", "created_at", "updated_at"}).
			AddRow(task.ID, task.TaskID, task.UserID, task.Type, task.TemplateID, task.Template, task.Amount, task.Status, task.CreatedAt, task.UpdatedAt))

	result, err := repo.GetTaskByID(context.Background(), 1)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, task.ID, result.ID)
	assert.Equal(t, task.TaskID, result.TaskID)
	assert.Equal(t, task.UserID, result.UserID)
	assert.Equal(t, task.Type, result.Type)
	assert.Equal(t, task.TemplateID, result.TemplateID)
	assert.Equal(t, task.Template, result.Template)
	assert.Equal(t, task.Amount, result.Amount)
	assert.Equal(t, task.Status, result.Status)

	require.NoError(t, mock.ExpectationsWereMet())
}

// TestGetTaskByID_DBError проверяет ошибку базы данных при получении задачи.
func TestGetTaskByID_DBError(t *testing.T) {
	repo, mock := setupTaskRepository(t)
	defer mock.Close()

	mock.ExpectQuery(`SELECT id, task_id, user_id, type, template_id, template, amount, status, created_at, updated_at`).
		WithArgs(int64(1)).
		WillReturnError(errors.New("db error"))

	result, err := repo.GetTaskByID(context.Background(), 1)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "db error")

	require.NoError(t, mock.ExpectationsWereMet())
}
