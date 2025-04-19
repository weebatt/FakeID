package kafka

import (
	"context"
	"fmt"
	"task-service/internal/config"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/sony/gobreaker"
	"go.uber.org/zap"
)

type KafkaProducer struct {
	Producer *kafka.Producer
	logger   *zap.SugaredLogger
	cb       *gobreaker.CircuitBreaker
	topic    string
}

func NewKafkaProducer(ctx context.Context, cfg config.KafkaConfig, logger *zap.SugaredLogger) (*KafkaProducer, error) {
	//Circuit Breaker
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

	var producer *kafka.Producer
	var err error

	for attempt := 1; attempt <= cfg.MaxRetries; attempt++ {
		// Проверяем отмену контекста перед попыткой
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("Kafka producer creation canceled: %w", ctx.Err())
		default:
		}

		producer, err = kafka.NewProducer(&kafka.ConfigMap{
			"bootstrap.servers": cfg.Brokers,
			"acks":              "all",
			"retries":           3,
			"retry.backoff.ms":  1000,
		})
		if err != nil {
			logger.Warnf("Failed to create Kafka producer (attempt %d): %v", attempt, err)
			if attempt == cfg.MaxRetries {
				return nil, fmt.Errorf("unable to initialize Kafka producer after %d attempts: %w", cfg.MaxRetries, err)
			}
			// Ожидаем перед следующей попыткой с учетом контекста
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("Kafka producer creation canceled: %w", ctx.Err())
			case <-time.After(time.Duration(cfg.RetryDelay) * time.Second):
			}
			continue
		}

		// Проверяем подключение
		_, err = producer.GetMetadata(nil, true, int(time.Duration(cfg.Timeout)*time.Second/time.Millisecond))
		if err == nil {
			logger.Infof("Kafka producer successfully connected on attempt %d", attempt)
			break
		}

		logger.Warnf("Kafka metadata check failed (attempt %d): %v", attempt, err)

		producer.Close()
		producer = nil

		if attempt < cfg.MaxRetries {
			// Ожидаем перед следующей попыткой с учетом контекста
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("Kafka producer creation canceled: %w", ctx.Err())
			case <-time.After(time.Duration(cfg.RetryDelay) * time.Second):
			}
		}
	}

	if producer == nil {
		return nil, fmt.Errorf("failed to establish Kafka connection after %d attempts", cfg.MaxRetries)
	}

	return &KafkaProducer{
		Producer: producer,
		logger:   logger,
		cb:       cb,
		topic:    cfg.Topic,
	}, nil
}

func (k *KafkaProducer) Produce(ctx context.Context, key []byte, value []byte) error {
	_, err := k.cb.Execute(func() (interface{}, error) {
		deliveryChan := make(chan kafka.Event)
		defer close(deliveryChan)

		err := k.Producer.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &k.topic, Partition: -1},
			Key:            key,
			Value:          value,
		}, deliveryChan)
		if err != nil {
			return nil, err
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case e := <-deliveryChan:
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					return nil, ev.TopicPartition.Error
				}
				return nil, nil
			default:
				return nil, fmt.Errorf("unexpected event type: %T", e)
			}
		}
	})
	if err != nil {
		k.logger.Errorf("Circuit Breaker rejected Produce: %v", err)
	}
	return err
}

func (k *KafkaProducer) Close() error {
	if k.Producer != nil {
		remaining := k.Producer.Flush(5000)
		k.Producer.Close()
		k.logger.Info("Kafka producer connection closed")
		if remaining > 0 {
			return fmt.Errorf("failed to flush %d messages before closing Kafka producer", remaining)
		}
	}
	return nil
}
