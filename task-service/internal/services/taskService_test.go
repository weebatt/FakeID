package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"task-service/internal/models"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// fakeTaskRepository — фейковая реализация TaskRepository.
type fakeTaskRepository struct {
	createNewTaskFunc func(ctx context.Context, task models.Task) (int64, error)
	getTaskByIDFunc   func(ctx context.Context, id int64) (*models.Task, error)
}

func (f *fakeTaskRepository) CreateNewTask(ctx context.Context, task models.Task) (int64, error) {
	return f.createNewTaskFunc(ctx, task)
}

func (f *fakeTaskRepository) GetTaskByID(ctx context.Context, id int64) (*models.Task, error) {
	return f.getTaskByIDFunc(ctx, id)
}

// fakeRedisClient — фейковая реализация RedisClient.
type fakeRedisClient struct {
	setFunc func(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	getFunc func(ctx context.Context, key string) (string, error)
}

func (f *fakeRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return f.setFunc(ctx, key, value, expiration)
}

func (f *fakeRedisClient) Get(ctx context.Context, key string) (string, error) {
	return f.getFunc(ctx, key)
}

func (f *fakeRedisClient) Close() error {
	return nil
}

// fakeKafkaProducer — фейковая реализация KafkaProducer.
type fakeKafkaProducer struct {
	produceFunc func(ctx context.Context, key, value []byte) error
	closeFunc   func() error
}

func (f *fakeKafkaProducer) Produce(ctx context.Context, key, value []byte) error {
	return f.produceFunc(ctx, key, value)
}

func (f *fakeKafkaProducer) Close() error {
	return f.closeFunc()
}

// fakeTemplateClient — фейковая реализация TemplateClient.
type fakeTemplateClient struct {
	doFunc func(req *http.Request) (*http.Response, error)
}

func (f *fakeTemplateClient) Do(req *http.Request) (*http.Response, error) {
	return f.doFunc(req)
}

// TestCreateNewTask_Success проверяет успешное создание задачи.
func TestCreateNewTask_Success(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	sugaredLogger := logger.Sugar()

	// Настраиваем фейковые зависимости
	repo := &fakeTaskRepository{
		createNewTaskFunc: func(ctx context.Context, task models.Task) (int64, error) {
			return 1, nil
		},
	}

	redisClient := &fakeRedisClient{
		setFunc: func(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
			return nil
		},
	}

	kafkaProducer := &fakeKafkaProducer{
		produceFunc: func(ctx context.Context, key, value []byte) error {
			return nil
		},
		closeFunc: func() error {
			return nil
		},
	}

	templateResponse := `{"id":"template-456","name":"Test Template","content":{"name":"{{name}}","age":"{{age}}"}}`
	templateClient := &fakeTemplateClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader([]byte(templateResponse))),
				Header:     make(http.Header),
			}, nil
		},
	}

	svc := &taskService{
		repo:           repo,
		redis:          redisClient,
		kafka:          kafkaProducer,
		logger:         sugaredLogger,
		templateClient: templateClient,
	}

	ctx := context.Background()
	task := models.Task{
		TaskID:     "task-123",
		TemplateID: "template-456",
		Amount:     100,
	}

	id, err := svc.CreateNewTask(ctx, task)
	require.NoError(t, err)
	assert.Equal(t, int64(1), id)
}

// TestCreateNewTask_TemplateClientError проверяет ошибку при запросе к template-service.
func TestCreateNewTask_TemplateClientError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	sugaredLogger := logger.Sugar()

	repo := &fakeTaskRepository{}
	redisClient := &fakeRedisClient{
		setFunc: func(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
			return nil
		},
	}
	kafkaProducer := &fakeKafkaProducer{}
	templateClient := &fakeTemplateClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return nil, errors.New("network error")
		},
	}

	svc := &taskService{
		repo:           repo,
		redis:          redisClient,
		kafka:          kafkaProducer,
		logger:         sugaredLogger,
		templateClient: templateClient,
	}

	ctx := context.Background()
	task := models.Task{
		TaskID:     "task-123",
		TemplateID: "template-456",
		Amount:     100,
	}

	id, err := svc.CreateNewTask(ctx, task)
	require.Error(t, err)
	assert.Equal(t, int64(0), id)
	assert.Contains(t, err.Error(), "network error")
}

// TestCreateNewTask_RepoError проверяет ошибку при сохранении задачи в репозиторий.
func TestCreateNewTask_RepoError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	sugaredLogger := logger.Sugar()

	repo := &fakeTaskRepository{
		createNewTaskFunc: func(ctx context.Context, task models.Task) (int64, error) {
			return 0, errors.New("repo error")
		},
	}
	redisClient := &fakeRedisClient{
		setFunc: func(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
			return nil
		},
	}
	kafkaProducer := &fakeKafkaProducer{}
	templateClient := &fakeTemplateClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader([]byte(`{"id":"template-456","name":"Test Template","content":{"name":"{{name}}","age":"{{age}}"}}`))),
				Header:     make(http.Header),
			}, nil
		},
	}

	svc := &taskService{
		repo:           repo,
		redis:          redisClient,
		kafka:          kafkaProducer,
		logger:         sugaredLogger,
		templateClient: templateClient,
	}

	ctx := context.Background()
	task := models.Task{
		TaskID:     "task-123",
		TemplateID: "template-456",
		Amount:     100,
	}

	id, err := svc.CreateNewTask(ctx, task)
	require.Error(t, err)
	assert.Equal(t, int64(0), id)
	assert.Contains(t, err.Error(), "repo error")
}

// TestCreateNewTask_KafkaError проверяет ошибку при отправке в Kafka.
func TestCreateNewTask_KafkaError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	sugaredLogger := logger.Sugar()

	repo := &fakeTaskRepository{
		createNewTaskFunc: func(ctx context.Context, task models.Task) (int64, error) {
			return 1, nil
		},
	}
	redisClient := &fakeRedisClient{
		setFunc: func(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
			return nil
		},
	}
	kafkaProducer := &fakeKafkaProducer{
		produceFunc: func(ctx context.Context, key, value []byte) error {
			return errors.New("kafka error")
		},
		closeFunc: func() error {
			return nil
		},
	}
	templateClient := &fakeTemplateClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader([]byte(`{"id":"template-456","name":"Test Template","content":{"name":"{{name}}","age":"{{age}}"}}`))),
				Header:     make(http.Header),
			}, nil
		},
	}

	svc := &taskService{
		repo:           repo,
		redis:          redisClient,
		kafka:          kafkaProducer,
		logger:         sugaredLogger,
		templateClient: templateClient,
	}

	ctx := context.Background()
	task := models.Task{
		TaskID:     "task-123",
		TemplateID: "template-456",
		Amount:     100,
	}

	id, err := svc.CreateNewTask(ctx, task)
	require.NoError(t, err) // Ошибка Kafka не влияет на результат
	assert.Equal(t, int64(1), id)
}

// TestCreateNewTask_RedisError проверяет ошибку при сохранении в Redis.
func TestCreateNewTask_RedisError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	sugaredLogger := logger.Sugar()

	repo := &fakeTaskRepository{
		createNewTaskFunc: func(ctx context.Context, task models.Task) (int64, error) {
			return 1, nil
		},
	}
	redisClient := &fakeRedisClient{
		setFunc: func(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
			return errors.New("redis error")
		},
	}
	kafkaProducer := &fakeKafkaProducer{
		produceFunc: func(ctx context.Context, key, value []byte) error {
			return nil
		},
		closeFunc: func() error {
			return nil
		},
	}
	templateClient := &fakeTemplateClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader([]byte(`{"id":"template-456","name":"Test Template","content":{"name":"{{name}}","age":"{{age}}"}}`))),
				Header:     make(http.Header),
			}, nil
		},
	}

	svc := &taskService{
		repo:           repo,
		redis:          redisClient,
		kafka:          kafkaProducer,
		logger:         sugaredLogger,
		templateClient: templateClient,
	}

	ctx := context.Background()
	task := models.Task{
		TaskID:     "task-123",
		TemplateID: "template-456",
		Amount:     100,
	}

	id, err := svc.CreateNewTask(ctx, task)
	require.NoError(t, err) // Ошибка Redis не влияет на результат
	assert.Equal(t, int64(1), id)
}

// TestGetTaskByID_FromCache проверяет получение задачи из кэша.
func TestGetTaskByID_FromCache(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	sugaredLogger := logger.Sugar()

	repo := &fakeTaskRepository{}
	redisClient := &fakeRedisClient{
		getFunc: func(ctx context.Context, key string) (string, error) {
			task := models.Task{
				ID:     1,
				TaskID: "task-123",
				Amount: 100,
			}
			data, _ := json.Marshal(task)
			return string(data), nil
		},
	}
	kafkaProducer := &fakeKafkaProducer{}
	templateClient := &fakeTemplateClient{}

	svc := &taskService{
		repo:           repo,
		redis:          redisClient,
		kafka:          kafkaProducer,
		logger:         sugaredLogger,
		templateClient: templateClient,
	}

	ctx := context.Background()
	task, err := svc.GetTaskByID(ctx, 1)
	require.NoError(t, err)
	require.NotNil(t, task)
	assert.Equal(t, int64(1), task.ID)
	assert.Equal(t, "task-123", task.TaskID)
}

// TestGetTaskByID_FromCache_InvalidData проверяет ошибку при некорректных данных в кэше.
func TestGetTaskByID_FromCache_InvalidData(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	sugaredLogger := logger.Sugar()

	repo := &fakeTaskRepository{
		getTaskByIDFunc: func(ctx context.Context, id int64) (*models.Task, error) {
			return &models.Task{
				ID:     1,
				TaskID: "task-123",
				Amount: 100,
			}, nil
		},
	}
	redisClient := &fakeRedisClient{
		getFunc: func(ctx context.Context, key string) (string, error) {
			return "invalid json", nil
		},
		setFunc: func(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
			return nil
		},
	}
	kafkaProducer := &fakeKafkaProducer{}
	templateClient := &fakeTemplateClient{}

	svc := &taskService{
		repo:           repo,
		redis:          redisClient,
		kafka:          kafkaProducer,
		logger:         sugaredLogger,
		templateClient: templateClient,
	}

	ctx := context.Background()
	task, err := svc.GetTaskByID(ctx, 1)
	require.NoError(t, err)
	require.NotNil(t, task)
	assert.Equal(t, int64(1), task.ID)
	assert.Equal(t, "task-123", task.TaskID)
}

// TestGetTaskByID_FromRepo проверяет получение задачи из репозитория.
func TestGetTaskByID_FromRepo(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	sugaredLogger := logger.Sugar()

	repo := &fakeTaskRepository{
		getTaskByIDFunc: func(ctx context.Context, id int64) (*models.Task, error) {
			return &models.Task{
				ID:     1,
				TaskID: "task-123",
				Amount: 100,
			}, nil
		},
	}
	redisClient := &fakeRedisClient{
		getFunc: func(ctx context.Context, key string) (string, error) {
			return "", errors.New("not found")
		},
		setFunc: func(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
			return nil
		},
	}
	kafkaProducer := &fakeKafkaProducer{}
	templateClient := &fakeTemplateClient{}

	svc := &taskService{
		repo:           repo,
		redis:          redisClient,
		kafka:          kafkaProducer,
		logger:         sugaredLogger,
		templateClient: templateClient,
	}

	ctx := context.Background()
	task, err := svc.GetTaskByID(ctx, 1)
	require.NoError(t, err)
	require.NotNil(t, task)
	assert.Equal(t, int64(1), task.ID)
	assert.Equal(t, "task-123", task.TaskID)
}

// TestGetTaskByID_RepoError проверяет ошибку при получении задачи из репозитория.
func TestGetTaskByID_RepoError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	sugaredLogger := logger.Sugar()

	repo := &fakeTaskRepository{
		getTaskByIDFunc: func(ctx context.Context, id int64) (*models.Task, error) {
			return nil, errors.New("repo error")
		},
	}
	redisClient := &fakeRedisClient{
		getFunc: func(ctx context.Context, key string) (string, error) {
			return "", errors.New("not found")
		},
	}
	kafkaProducer := &fakeKafkaProducer{}
	templateClient := &fakeTemplateClient{}

	svc := &taskService{
		repo:           repo,
		redis:          redisClient,
		kafka:          kafkaProducer,
		logger:         sugaredLogger,
		templateClient: templateClient,
	}

	ctx := context.Background()
	task, err := svc.GetTaskByID(ctx, 1)
	require.Error(t, err)
	assert.Nil(t, task)
	assert.Contains(t, err.Error(), "repo error")
}
