package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNew_Success проверяет успешную загрузку конфигурации.
func TestNew_Success(t *testing.T) {
	// Устанавливаем переменные окружения
	os.Setenv("ENV", "prod")
	os.Setenv("HOST", "127.0.0.1")
	os.Setenv("PORT", "8081")
	os.Setenv("MAX_RETRIES", "3")
	os.Setenv("RETRY_DELAY", "2")
	os.Setenv("PG_HOST", "pg-host")
	os.Setenv("PG_PORT", "5433")
	os.Setenv("PG_USER", "pg-user")
	os.Setenv("PG_PASSWORD", "pg-pass")
	os.Setenv("PG_DBNAME", "pg-db")
	os.Setenv("PG_SSLMODE", "require")
	os.Setenv("PG_MAX_CONNS", "100")
	os.Setenv("PG_MIN_CONNS", "20")
	os.Setenv("PG_TIMEOUT", "10")
	os.Setenv("PG_MAX_RETRIES", "4")
	os.Setenv("PG_RETRY_DELAY", "3")
	os.Setenv("REDIS_HOST", "redis-host")
	os.Setenv("REDIS_PORT", "6380")
	os.Setenv("REDIS_PASSWORD", "redis-pass")
	os.Setenv("REDIS_DB", "1")
	os.Setenv("REDIS_TIMEOUT", "6")
	os.Setenv("REDIS_MAX_RETRIES", "3")
	os.Setenv("REDIS_RETRY_DELAY", "4")

	// Вызываем New()
	cfg, err := New()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Проверяем значения
	assert.Equal(t, "prod", cfg.Env)
	// HTTP Server
	assert.Equal(t, "127.0.0.1", cfg.HTTPServer.Host)
	assert.Equal(t, "8081", cfg.HTTPServer.Port)
	assert.Equal(t, 3, cfg.HTTPServer.MaxRetries)
	assert.Equal(t, 2, cfg.HTTPServer.RetryDelay)
	// Postgres
	assert.Equal(t, "pg-host", cfg.Postgres.Host)
	assert.Equal(t, "5433", cfg.Postgres.Port)
	assert.Equal(t, "pg-user", cfg.Postgres.User)
	assert.Equal(t, "pg-pass", cfg.Postgres.Password)
	assert.Equal(t, "pg-db", cfg.Postgres.DBName)
	assert.Equal(t, "require", cfg.Postgres.SSLMode)
	assert.Equal(t, int32(100), cfg.Postgres.MaxConns)
	assert.Equal(t, int32(20), cfg.Postgres.MinConns)
	assert.Equal(t, 10, cfg.Postgres.Timeout)
	assert.Equal(t, 4, cfg.Postgres.MaxRetries)
	assert.Equal(t, 3, cfg.Postgres.RetryDelay)
	// Redis
	assert.Equal(t, "redis-host", cfg.Redis.Host)
	assert.Equal(t, "6380", cfg.Redis.Port)
	assert.Equal(t, "redis-pass", cfg.Redis.Password)
	assert.Equal(t, 1, cfg.Redis.DB)
	assert.Equal(t, 6, cfg.Redis.Timeout)
	assert.Equal(t, 3, cfg.Redis.MaxRetries)
	assert.Equal(t, 4, cfg.Redis.RetryDelay)

	// Очищаем переменные окружения после теста
	os.Clearenv()
}

// TestNew_DefaultValues проверяет использование значений по умолчанию.
func TestNew_DefaultValues(t *testing.T) {
	// Устанавливаем только обязательные переменные (остальные должны быть значениями по умолчанию)
	os.Setenv("HOST", "localhost")
	os.Setenv("PG_HOST", "localhost")
	os.Setenv("PG_DBNAME", "postgres")
	os.Setenv("REDIS_HOST", "localhost")

	cfg, err := New()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Проверяем значения по умолчанию
	assert.Equal(t, "prod", cfg.Env)
	// HTTP Server
	assert.Equal(t, "localhost", cfg.HTTPServer.Host)
	assert.Equal(t, "8080", cfg.HTTPServer.Port)
	assert.Equal(t, 5, cfg.HTTPServer.MaxRetries)
	assert.Equal(t, 5, cfg.HTTPServer.RetryDelay)
	// Postgres
	assert.Equal(t, "localhost", cfg.Postgres.Host)
	assert.Equal(t, "5432", cfg.Postgres.Port)
	assert.Equal(t, "postgres", cfg.Postgres.User)
	assert.Equal(t, "", cfg.Postgres.Password)
	assert.Equal(t, "postgres", cfg.Postgres.DBName)
	assert.Equal(t, "disable", cfg.Postgres.SSLMode)
	assert.Equal(t, int32(50), cfg.Postgres.MaxConns)
	assert.Equal(t, int32(10), cfg.Postgres.MinConns)
	assert.Equal(t, 5, cfg.Postgres.Timeout)
	assert.Equal(t, 5, cfg.Postgres.MaxRetries)
	assert.Equal(t, 2, cfg.Postgres.RetryDelay)
	// Redis
	assert.Equal(t, "localhost", cfg.Redis.Host)
	assert.Equal(t, "6379", cfg.Redis.Port)
	assert.Equal(t, "", cfg.Redis.Password)
	assert.Equal(t, 0, cfg.Redis.DB)
	assert.Equal(t, 5, cfg.Redis.Timeout)
	assert.Equal(t, 5, cfg.Redis.MaxRetries)
	assert.Equal(t, 3, cfg.Redis.RetryDelay)

	os.Clearenv()
}

// TestNew_ValidationError_Env проверяет ошибку валидации для поля Env.
func TestNew_ValidationError_Env(t *testing.T) {
	// Устанавливаем некорректное значение для Env
	os.Setenv("ENV", "invalid")
	os.Setenv("HOST", "localhost")
	os.Setenv("PG_HOST", "localhost")
	os.Setenv("PG_DBNAME", "postgres")
	os.Setenv("REDIS_HOST", "localhost")

	cfg, err := New()
	require.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "failed to validate config")
	assert.Contains(t, err.Error(), "Env")

	os.Clearenv()
}

// TestNew_ValidationError_SSLMode проверяет ошибку валидации для поля SSLMode.
func TestNew_ValidationError_SSLMode(t *testing.T) {
	// Устанавливаем некорректное значение для PG_SSLMODE
	os.Setenv("PG_SSLMODE", "invalid")
	os.Setenv("HOST", "localhost")
	os.Setenv("PG_HOST", "localhost")
	os.Setenv("PG_DBNAME", "postgres")
	os.Setenv("REDIS_HOST", "localhost")

	cfg, err := New()
	require.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "failed to validate config")
	assert.Contains(t, err.Error(), "SSLMode")

	os.Clearenv()
}

// TestNew_ValidationError_PortNumeric проверяет ошибку валидации для поля Port (должно быть числом).
func TestNew_ValidationError_PortNumeric(t *testing.T) {
	// Устанавливаем некорректное значение для PORT
	os.Setenv("PORT", "invalid")
	os.Setenv("HOST", "localhost")
	os.Setenv("PG_HOST", "localhost")
	os.Setenv("PG_DBNAME", "postgres")
	os.Setenv("REDIS_HOST", "localhost")

	cfg, err := New()
	require.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "failed to validate config")
	assert.Contains(t, err.Error(), "Port")

	os.Clearenv()
}

// TestNew_MissingRequiredField проверяет ошибку при отсутствии обязательного поля.
func TestNew_MissingRequiredField(t *testing.T) {
	// Не устанавливаем обязательное поле HOST
	os.Setenv("PG_HOST", "localhost")
	os.Setenv("PG_DBNAME", "postgres")
	os.Setenv("REDIS_HOST", "localhost")

	cfg, err := New()
	require.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "failed to validate config")
	assert.Contains(t, err.Error(), "Host")

	os.Clearenv()
}
