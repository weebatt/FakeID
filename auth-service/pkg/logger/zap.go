// logger/zap.go
package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	*zap.Logger
}

// NewLogger создает новый экземпляр логгера с заданной конфигурацией
func NewLogger(environment string) (*Logger, error) {
	var cfg zap.Config
	var logger *zap.Logger
	var err error

	if environment == "production" {
		cfg = zap.NewProductionConfig()
		cfg.EncoderConfig.TimeKey = "timestamp"
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		cfg.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	} else {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		cfg.EncoderConfig.TimeKey = "timestamp"
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	// Добавляем caller и stacktrace для ошибок
	cfg.EncoderConfig.CallerKey = "caller"
	cfg.EncoderConfig.StacktraceKey = "stacktrace"

	logger, err = cfg.Build(zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	if err != nil {
		return nil, err
	}

	return &Logger{logger}, nil
}

// Sync вызывает метод Sync базового логгера
func (l *Logger) Sync() error {
	return l.Logger.Sync()
}

// WithFields добавляет дополнительные поля к логгеру
func (l *Logger) WithFields(fields ...interface{}) *Logger {
	zapFields := make([]zap.Field, 0, len(fields)/2)
	for i := 0; i < len(fields)-1; i += 2 {
		if key, ok := fields[i].(string); ok {
			zapFields = append(zapFields, zap.Any(key, fields[i+1]))
		}
	}
	return &Logger{l.Logger.With(zapFields...)}
}

// Добавляем методы для удобства использования
func (l *Logger) Debug(msg string, fields ...interface{}) {
	l.WithFields(fields...).Logger.Debug(msg)
}

func (l *Logger) Info(msg string, fields ...interface{}) {
	l.WithFields(fields...).Logger.Info(msg)
}

func (l *Logger) Warn(msg string, fields ...interface{}) {
	l.WithFields(fields...).Logger.Warn(msg)
}

func (l *Logger) Error(msg string, fields ...interface{}) {
	l.WithFields(fields...).Logger.Error(msg)
}

func (l *Logger) Fatal(msg string, fields ...interface{}) {
	l.WithFields(fields...).Logger.Fatal(msg)
}
