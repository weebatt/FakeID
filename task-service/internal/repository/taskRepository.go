package repository

import (
	"context"
	"task-service/internal/models"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

// DB — интерфейс для методов базы данных, используемых postgresTaskRepository.
type DB interface {
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Exec(ctx context.Context, sql string, args ...interface{}) error
	Ping(ctx context.Context) error
	Close()
}

type TaskRepository interface {
	CreateNewTask(ctx context.Context, task models.Task) (int64, error)
	GetTaskByID(ctx context.Context, id int64) (*models.Task, error)
}

type postgresTaskRepository struct {
	db     DB
	logger *zap.SugaredLogger
}

func NewTaskRepository(db DB, logger *zap.SugaredLogger) *postgresTaskRepository {
	return &postgresTaskRepository{db: db, logger: logger}
}

func (r *postgresTaskRepository) CreateNewTask(ctx context.Context, task models.Task) (int64, error) {
	task.TaskID = uuid.NewString()
	task.Status = "pending"
	task.CreatedAt = time.Now()
	task.UpdatedAt = task.CreatedAt

	query := `INSERT INTO tasks (task_id, user_id, type, template_id, template, amount, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id`

	var id int64
	err := r.db.QueryRow(ctx, query, task.TaskID, task.UserID, task.Type, task.TemplateID, task.Template, task.Amount, task.Status, task.CreatedAt, task.UpdatedAt).Scan(&id)
	if err != nil {
		r.logger.Errorf("Failed to insert task: %v", err)
		return 0, err
	}

	r.logger.Infof("Task created with ID: %d", id)
	return id, nil
}

func (r *postgresTaskRepository) GetTaskByID(ctx context.Context, id int64) (*models.Task, error) {
	query := `SELECT id, task_id, user_id, type, template_id, template, amount, status, created_at, updated_at FROM tasks WHERE id = $1`

	var task models.Task
	err := r.db.QueryRow(ctx, query, id).Scan(&task.ID, &task.TaskID, &task.UserID, &task.Type, &task.TemplateID, &task.Template, &task.Amount, &task.Status, &task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		r.logger.Errorf("Failed to get task: %v", err)
		return nil, err
	}

	r.logger.Infof("Task retrieved with ID: %d", id)
	return &task, nil
}
