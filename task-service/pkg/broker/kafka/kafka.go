package kafka

import (
	"context"
	"fmt"
	"strings"
	"task-service/internal/config"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/sony/gobreaker"
	"go.uber.org/zap"
)

// writerInterface defines the methods of kafka.Writer that we need.
type writerInterface interface {
	WriteMessages(ctx context.Context, messages ...kafka.Message) error
	Close() error
}

// KafkaProducer is an interface for producing messages to Kafka.
type KafkaProducer interface {
	Produce(ctx context.Context, key []byte, value []byte) error
	Close() error
}

type kafkaProducer struct {
	writer writerInterface
	logger *zap.SugaredLogger
	cb     *gobreaker.CircuitBreaker
	topic  string
}

func NewKafkaProducer(ctx context.Context, cfg config.KafkaConfig, logger *zap.SugaredLogger) (KafkaProducer, error) {
	// Circuit Breaker
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "kafka-producer",
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

	var writer writerInterface
	brokerList := strings.Split(cfg.Brokers, ",")

	for attempt := 1; attempt <= cfg.MaxRetries; attempt++ {
		// Wait for context cancellation
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("Kafka producer creation canceled: %w", ctx.Err())
		default:
		}

		// Create a new Kafka writer
		writer = &kafka.Writer{
			Addr:                   kafka.TCP(brokerList...),
			Topic:                  cfg.Topic,
			Balancer:               &kafka.LeastBytes{},
			RequiredAcks:           kafka.RequireAll, // Equivalent to "acks": "all"
			MaxAttempts:            3,                // Equivalent to "retries": 3
			BatchTimeout:           1 * time.Second,  // Equivalent to "retry.backoff.ms": 1000
			AllowAutoTopicCreation: true,
		}

		// Check connection by attempting to fetch metadata
		conn, err := kafka.Dial("tcp", brokerList[0])
		if err != nil {
			logger.Warnf("Failed to create Kafka producer (attempt %d): %v", attempt, err)
			if attempt == cfg.MaxRetries {
				return nil, fmt.Errorf("unable to initialize Kafka producer after %d attempts: %w", cfg.MaxRetries, err)
			}
			// Check for context cancellation
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("Kafka producer creation canceled: %w", ctx.Err())
			case <-time.After(time.Duration(cfg.RetryDelay) * time.Second):
			}
			continue
		}

		// Fetch metadata to verify connection
		_, err = conn.ReadPartitions(cfg.Topic)
		conn.Close()
		if err == nil {
			logger.Infof("Kafka producer successfully connected on attempt %d", attempt)
			break
		}

		logger.Warnf("Kafka metadata check failed (attempt %d): %v", attempt, err)

		writer.Close()
		writer = nil

		if attempt < cfg.MaxRetries {
			// Check for context cancellation
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("Kafka producer creation canceled: %w", ctx.Err())
			case <-time.After(time.Duration(cfg.RetryDelay) * time.Second):
			}
		}
	}

	if writer == nil {
		return nil, fmt.Errorf("failed to establish Kafka connection after %d attempts", cfg.MaxRetries)
	}

	return &kafkaProducer{
		writer: writer,
		logger: logger,
		cb:     cb,
		topic:  cfg.Topic,
	}, nil
}

func (k *kafkaProducer) Produce(ctx context.Context, key []byte, value []byte) error {
	_, err := k.cb.Execute(func() (interface{}, error) {
		err := k.writer.WriteMessages(ctx, kafka.Message{
			Key:   key,
			Value: value,
		})
		if err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		k.logger.Errorf("Circuit Breaker rejected Produce: %v", err)
	}
	return err
}

func (k *kafkaProducer) Close() error {
	if k.writer != nil {
		err := k.writer.Close()
		k.logger.Info("Kafka producer connection closed")
		return err
	}
	return nil
}
