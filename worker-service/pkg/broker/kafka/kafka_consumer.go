package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"worker-service/internal/config"
	"worker-service/internal/models"

	"github.com/segmentio/kafka-go"
	"github.com/sony/gobreaker"
	"go.uber.org/zap"
)

// KafkaConsumer defines the interface for consuming messages from Kafka
type KafkaConsumer interface {
	Consume(ctx context.Context) error
	Close() error
}

type kafkaConsumer struct {
	reader *kafka.Reader
	logger *zap.SugaredLogger
	cb     *gobreaker.CircuitBreaker
	config config.KafkaConfig
}

// NewKafkaConsumer creates a new Kafka consumer instance
func NewKafkaConsumer(cfg config.KafkaConfig, logger *zap.SugaredLogger) (KafkaConsumer, error) {
	// Circuit Breaker configuration
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "kafka-consumer",
		MaxRequests: 1,
		Interval:    30 * time.Second,
		Timeout:     time.Duration(cfg.Timeout) * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 3
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			logger.Infof("circuit breaker %s state changed from %s to %s", name, from.String(), to.String())
		},
	})

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     strings.Split(cfg.Brokers, ","),
		Topic:       cfg.Topic,
		GroupID:     "worker-service-group",
		MinBytes:    1,
		MaxBytes:    10e6,
		MaxWait:     1 * time.Second,
		StartOffset: kafka.FirstOffset,
	})

	return &kafkaConsumer{
		reader: reader,
		logger: logger,
		cb:     cb,
		config: cfg,
	}, nil
}

func (k *kafkaConsumer) Consume(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			k.logger.Info("Stopping Kafka consumer due to context cancellation")
			return nil
		default:
			// Use circuit breaker for message consumption
			_, err := k.cb.Execute(func() (interface{}, error) {
				msg, err := k.reader.ReadMessage(ctx)
				if err != nil {
					return nil, fmt.Errorf("failed to read message: %w", err)
				}

				var task models.Task
				if err := json.Unmarshal(msg.Value, &task); err != nil {
					k.logger.Errorf("Failed to unmarshal task: %v", err)
					return nil, nil // Skip bad messages
				}

				k.logger.Infof("Я ПРИНЯЛ ТАСКУ [ID: %s], ГОТОВ К ГЕНЕРАЦИИ ДАННЫХ", task.ID)
				return nil, nil
			})

			if err != nil {
				k.logger.Errorf("Error consuming message: %v", err)
				time.Sleep(time.Duration(k.config.RetryDelay) * time.Second)
			}
		}
	}
}

func (k *kafkaConsumer) Close() error {
	if k.reader != nil {
		if err := k.reader.Close(); err != nil {
			k.logger.Errorf("Failed to close Kafka reader: %v", err)
			return err
		}
		k.logger.Info("Kafka consumer connection closed")
	}
	return nil
}
