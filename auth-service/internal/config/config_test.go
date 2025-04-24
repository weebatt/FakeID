package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLoad_Defaults(t *testing.T) {
	// без переменных окружения должны подставиться дефолты
	cfg, err := Load()
	require.NoError(t, err)

	require.Equal(t, "8080", cfg.Server.Port)
	require.Equal(t, "0.0.0.0", cfg.Server.Host)
	require.Equal(t, 15*time.Second, cfg.Server.ReadTimeout)
	require.Equal(t, "localhost", cfg.Database.Host)
	require.Equal(t, "your-secret-key", cfg.JWT.Secret)
}

func TestLoad_WithEnv(t *testing.T) {
	t.Setenv("SERVER_PORT", "9000")
	t.Setenv("DB_HOST", "postgres.internal")
	t.Setenv("JWT_SECRET", "super‑secret")

	cfg, err := Load()
	require.NoError(t, err)

	require.Equal(t, "9000", cfg.Server.Port)
	require.Equal(t, "postgres.internal", cfg.Database.Host)
	require.Equal(t, "super‑secret", cfg.JWT.Secret)
}
