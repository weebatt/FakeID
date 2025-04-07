package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Task struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TaskID     string             `bson:"task_id" json:"task_id"` // UUID
	UserID     string             `bson:"user_id" json:"user_id"`
	Type       string             `bson:"type" json:"type"` //e.g. user/payment
	TemplateID *string            `bson:"template_id" json:"template_id"`
	Template   bson.M             `bson:"template" json:"template"`
	Amount     int                `bson:"amount" json:"amount"`
	Format     string             `bson:"format" json:"format"`
	Status     string             `bson:"status" json:"status"` //e.g. pending/done
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt  time.Time          `bson:"updated_at" json:"updated_at"`
}

type CreateTaskRequest struct {
	Type       string  `json:"type"`
	TemplateID *string `json:"template_id,omitempty"`
	Template   *bson.M `json:"template,omitempty"`
	Amount     int     `json:"amount"`
	Format     string  `json:"format"`
}
