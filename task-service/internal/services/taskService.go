package services

import (
	"context"
	"encoding/json"
	"strconv"
	"task-service/internal/models"
	"task-service/internal/repository"
	"task-service/pkg/broker/kafka"
	"task-service/pkg/db/redis"
	"time"

	"go.uber.org/zap"
)

type TaskService interface {
	CreateNewTask(ctx context.Context, task models.Task) (int64, error)
	GetTaskByID(ctx context.Context, id int64) (*models.Task, error)
}

type taskService struct {
	repo   repository.TaskRepository
	redis  *redis.Redis
	kafka  *kafka.KafkaProducer
	logger *zap.SugaredLogger
}

func NewTaskService(repo repository.TaskRepository, redis *redis.Redis, kafka *kafka.KafkaProducer, logger *zap.SugaredLogger) TaskService {
	return &taskService{
		repo:   repo,
		redis:  redis,
		kafka:  kafka,
		logger: logger,
	}
}

func (t *taskService) CreateNewTask(ctx context.Context, task models.Task) (int64, error) {
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
