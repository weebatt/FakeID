package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"task-service/internal/models"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

// DB — интерфейс для методов базы данных, используемых postgresTaskRepository.
type DB interface {
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, args ...interface{}) error
	Ping(ctx context.Context) error
	Close()
}

type TaskRepository interface {
	CreateNewTask(ctx context.Context, task models.Task) (int64, error)
	GetTaskByID(ctx context.Context, id int64) (*models.Task, error)
	ListTasks(ctx context.Context, filter models.TaskFilter) ([]models.Task, error)
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

func (r *postgresTaskRepository) ListTasks(ctx context.Context, filter models.TaskFilter) ([]models.Task, error) {
	query := `SELECT id, task_id, user_id, type, template_id, template, amount, status, created_at, updated_at 
              FROM tasks WHERE 1=1`

	args := make([]interface{}, 0)
	argCount := 1

	// Add filters if provided
	if filter.UserID != "" {
		query += fmt.Sprintf(" AND user_id = $%d", argCount)
		args = append(args, filter.UserID)
		argCount++
	}

	if filter.Type != "" {
		query += fmt.Sprintf(" AND type = $%d", argCount)
		args = append(args, filter.Type)
		argCount++
	}

	if filter.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, filter.Status)
		argCount++
	}

	// Add ordering and pagination
	query += " ORDER BY created_at DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, filter.Limit)
		argCount++

		if filter.Page > 0 {
			offset := (filter.Page - 1) * filter.Limit
			query += fmt.Sprintf(" OFFSET $%d", argCount)
			args = append(args, offset)
		}
	}

	// Execute the query with proper typing
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		r.logger.Errorf("Failed to query tasks: %v", err)
		return nil, err
	}
	defer rows.Close()

	// Parse the results into a properly typed slice
	tasks := []models.Task{}
	for rows.Next() {
		var task models.Task
		var templateBytes []byte // Use byte slice for JSONB data

		if err := rows.Scan(
			&task.ID,
			&task.TaskID,
			&task.UserID,
			&task.Type,
			&task.TemplateID,
			&templateBytes, // Scan JSONB as bytes
			&task.Amount,
			&task.Status,
			&task.CreatedAt,
			&task.UpdatedAt,
		); err != nil {
			r.logger.Errorf("Failed to scan task row: %v", err)
			return nil, err
		}

		// Parse the template JSON
		if len(templateBytes) > 0 {
			if err := json.Unmarshal(templateBytes, &task.Template); err != nil {
				r.logger.Warnf("Failed to parse template JSON for task %d: %v", task.ID, err)
			}
		}

		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		r.logger.Errorf("Error during rows iteration: %v", err)
		return nil, err
	}

	r.logger.Infof("Retrieved %d tasks", len(tasks))
	return tasks, nil
}
