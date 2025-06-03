package handlers

import (
	"auth-service/internal/utils"
	database "auth-service/pkg/db/postgres"
	"auth-service/pkg/db/redis"
	"auth-service/pkg/logger"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

type MockLogger struct {
	Calls []LogCall
}

type LogCall struct {
	Level string
	Msg   string
	Args  []interface{}
}

func (m *MockLogger) Info(msg string, keysAndValues ...interface{}) {
	m.Calls = append(m.Calls, LogCall{Level: "Info", Msg: msg, Args: keysAndValues})
}

func (m *MockLogger) Warn(msg string, keysAndValues ...interface{}) {
	m.Calls = append(m.Calls, LogCall{Level: "Warn", Msg: msg, Args: keysAndValues})
}

func (m *MockLogger) Error(msg string, keysAndValues ...interface{}) {
	m.Calls = append(m.Calls, LogCall{Level: "Error", Msg: msg, Args: keysAndValues})
}

func setupAuth(t *testing.T) (*AuthHandler, sqlmock.Sqlmock, *echo.Echo, *MockLogger) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)

	mock.ExpectPing()

	mockLogger := &MockLogger{}
	logger := &logger.Logger{}

	db := &database.Database{DB: sqlDB}
	rd, _ := redis.NewRedis("localhost:6379", "", 0) // real Redis не нужен: только client.Set/Del/Get
	ah := NewAuthHandler(db, rd, []byte("secret"), time.Hour, logger)

	e := echo.New()
	return ah, mock, e, mockLogger
}

func TestRegister_Success(t *testing.T) {
	ah, mock, e, _ := setupAuth(t)

	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs("alice@mail.com").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	mock.ExpectBegin()

	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs("alice@mail.com", sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(42))

	mock.ExpectCommit()

	reqBody := `{"email":"alice@mail.com","password":"P@ssw0rd"}`
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	require.NoError(t, ah.Register(c))
	require.Equal(t, http.StatusCreated, rec.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.EqualValues(t, 42, resp["user_id"])

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestLogin_Success(t *testing.T) {
	ah, mock, e, _ := setupAuth(t)

	plaintext := "P@ssw0rd!"
	hash, _ := utils.HashPassword(plaintext)

	// SELECT user
	mock.ExpectQuery(`SELECT id, email, password_hash`).
		WithArgs("bob@mail.com").
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash"}).
			AddRow(7, "bob@mail.com", hash))

	reqBody := `{"email":"bob@mail.com","password":"P@ssw0rd!"}`
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	require.NoError(t, ah.Login(c))
	require.Equal(t, http.StatusOK, rec.Code)

	var resp struct {
		Token string `json:"token"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))

	// токен должен валидироваться на том же секрете
	tk, err := jwt.Parse(resp.Token, func(t *jwt.Token) (interface{}, error) { return []byte("secret"), nil })
	require.NoError(t, err)
	require.True(t, tk.Valid)

	require.NoError(t, mock.ExpectationsWereMet())
}
