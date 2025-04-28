package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"task-service/internal/models"
	"task-service/internal/repository"
	"task-service/pkg/broker/kafka"
	"time"

	"go.uber.org/zap"
)

// Добавляем интерфейсы для зависимостей
type RedisClient interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Close() error
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type TaskService interface {
	CreateNewTask(ctx context.Context, task models.Task) (int64, error)
	GetTaskByID(ctx context.Context, id int64) (*models.Task, error)
	ListTasks(ctx context.Context, filter models.TaskFilter) ([]models.Task, error)
}

type taskService struct {
	repo           repository.TaskRepository
	redis          RedisClient
	kafka          kafka.KafkaProducer
	logger         *zap.SugaredLogger
	templateClient HTTPClient
}

func NewTaskService(
	repo repository.TaskRepository,
	redis RedisClient,
	kafka kafka.KafkaProducer,
	logger *zap.SugaredLogger,
	templateClient HTTPClient,
) TaskService {
	return &taskService{
		repo:           repo,
		redis:          redis,
		kafka:          kafka,
		logger:         logger,
		templateClient: templateClient,
	}
}

func (t *taskService) CreateNewTask(ctx context.Context, task models.Task) (int64, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("http://template-service:8082/template/%s", task.TemplateID), nil)
	if err != nil {
		t.logger.Errorf("Failed to create request to template-service: %v", err)
		return 0, err
	}

	resp, err := t.templateClient.Do(req)
	if err != nil {
		t.logger.Errorf("Failed to fetch template %s: %v", task.TemplateID, err)
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.logger.Errorf("Template %s not found, status: %d", task.TemplateID, resp.StatusCode)
		return 0, fmt.Errorf("template not found")
	}

	var template struct {
		ID      string                 `json:"id"`
		Name    string                 `json:"name"`
		Content map[string]interface{} `json:"content"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&template); err != nil {
		t.logger.Errorf("Failed to decode template response: %v", err)
		return 0, err
	}

	task.Template = template.Content

	id, err := t.repo.CreateNewTask(ctx, task)
	if err != nil {
		t.logger.Errorf("Failed to create task: %v", err)
		return 0, err
	}

	t.logger.Infof("Task created with ID: %d", id)

	task.ID = id
	taskData, err := json.Marshal(task)
	if err != nil {
		t.logger.Errorf("Failed to marshal task %d: %v", id, err)
		return id, nil
	}

	if err := t.redis.Set(ctx, "task:"+strconv.FormatInt(id, 10), taskData, time.Hour); err != nil {
		t.logger.Warnf("Failed to cache task %d: %v", id, err)
	}

	err = t.kafka.Produce(ctx, []byte(task.TaskID), taskData)
	if err != nil {
		t.logger.Errorf("Failed to send task %d to Kafka: %v", id, err)
		return id, nil
	}

	t.logger.Infof("Task %d sent to Kafka", id)
	return id, nil
}

func (t *taskService) GetTaskByID(ctx context.Context, id int64) (*models.Task, error) {
	cacheKey := "task:" + strconv.FormatInt(id, 10)
	taskData, err := t.redis.Get(ctx, cacheKey)
	if err == nil {
		var task models.Task
		if err := json.Unmarshal([]byte(taskData), &task); err == nil {
			t.logger.Debugf("Task %d found in Redis", id)
			return &task, nil
		}
	}

	task, err := t.repo.GetTaskByID(ctx, id)
	if err != nil {
		t.logger.Errorf("Failed to get task: %v", err)
		return nil, err
	}

	taskDataBytes, err := json.Marshal(task)
	if err != nil {
		t.logger.Warnf("Failed to marshal task %d for caching: %v", id, err)
	} else {
		if err := t.redis.Set(ctx, cacheKey, taskDataBytes, time.Hour); err != nil {
			t.logger.Warnf("Failed to cache task %d: %v", id, err)
		}
	}

	t.logger.Infof("Task retrieved with ID: %d", id)
	return task, nil
}

func (t *taskService) ListTasks(ctx context.Context, filter models.TaskFilter) ([]models.Task, error) {
	// First try to get cached results if the filter is simple
	if filter.IsCacheable() {
		cacheKey := fmt.Sprintf("tasks:%s:%s:%s:%d:%d",
			filter.UserID, filter.Type, filter.Status, filter.Page, filter.Limit)

		cachedData, err := t.redis.Get(ctx, cacheKey)
		if err == nil {
			var tasks []models.Task
			if err := json.Unmarshal([]byte(cachedData), &tasks); err == nil {
				t.logger.Debugf("Tasks found in Redis cache for key: %s", cacheKey)
				return tasks, nil
			}
		}
	}

	// If not in cache or not cacheable, get from database
	tasks, err := t.repo.ListTasks(ctx, filter)
	if err != nil {
		t.logger.Errorf("Failed to list tasks: %v", err)
		return nil, err
	}

	// Cache results if appropriate
	if filter.IsCacheable() && len(tasks) > 0 {
		cacheKey := fmt.Sprintf("tasks:%s:%s:%s:%d:%d",
			filter.UserID, filter.Type, filter.Status, filter.Page, filter.Limit)

		tasksData, err := json.Marshal(tasks)
		if err == nil {
			if err := t.redis.Set(ctx, cacheKey, tasksData, 5*time.Minute); err != nil {
				t.logger.Warnf("Failed to cache tasks: %v", err)
			}
		}
	}

	t.logger.Infof("Retrieved %d tasks", len(tasks))
	return tasks, nil
}
