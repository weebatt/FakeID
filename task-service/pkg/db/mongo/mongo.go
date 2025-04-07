package mongo

import (
	"context"
	"task-service/internal/config"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

func New(mongoConfig config.MongoConfig) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

}
