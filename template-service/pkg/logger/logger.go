package logger

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	*zap.SugaredLogger
}

func New(env string) (*Logger, error) {
	var cfg zap.Config

	switch env {
	case "dev":
		cfg = zap.Config{
			Level:            zap.NewAtomicLevelAt(zap.DebugLevel),
			Development:      true,
			Encoding:         "console",
			EncoderConfig:    zap.NewDevelopmentEncoderConfig(),
			OutputPaths:      []string{"stdout"},
			ErrorOutputPaths: []string{"stderr"},
		}
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	case "prod":
		cfg = zap.Config{
			Level:            zap.NewAtomicLevelAt(zap.InfoLevel),
			Development:      false,
			Encoding:         "json",
			EncoderConfig:    zap.NewProductionEncoderConfig(),
			OutputPaths:      []string{"stdout"},
			ErrorOutputPaths: []string{"stderr"},
		}
		cfg.EncoderConfig.TimeKey = "timestamp"
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	case "test":
		cfg = zap.Config{
			Level:            zap.NewAtomicLevelAt(zap.WarnLevel),
			Development:      true,
			Encoding:         "console",
			EncoderConfig:    zap.NewDevelopmentEncoderConfig(),
			OutputPaths:      []string{"stdout"},
			ErrorOutputPaths: []string{"stderr"},
		}
	default:
		return nil, fmt.Errorf("unknown environment: %s", env)
	}

	cfg.EncoderConfig.CallerKey = "caller"
	cfg.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	logger, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	return &Logger{logger.Sugar()}, nil
}

func (l *Logger) Sync() error {
	return l.SugaredLogger.Desugar().Sync()
}

func (l *Logger) With(args ...interface{}) *Logger {
	return &Logger{l.SugaredLogger.With(args...)}
}

func (l *Logger) Info(args ...interface{}) {
	l.SugaredLogger.Info(args...)
}

func (l *Logger) Infof(template string, args ...interface{}) {
	l.SugaredLogger.Infof(template, args...)
}

func (l *Logger) Fatal(args ...interface{}) {
	l.SugaredLogger.Fatal(args...)
	os.Exit(1)
}

func (l *Logger) Fatalf(template string, args ...interface{}) {
	l.SugaredLogger.Fatalf(template, args...)
	os.Exit(1)
}
