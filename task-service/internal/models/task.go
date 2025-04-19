package models

import (
	"time"
)

type Task struct {
	ID         int64                  `json:"id" db:"id"`
	TaskID     string                 `json:"task_id" db:"task_id"`
	UserID     string                 `json:"user_id" db:"user_id"`
	Type       string                 `json:"type" db:"type"`
	TemplateID string                 `json:"template_id" db:"template_id"`
	Template   map[string]interface{} `json:"template" db:"template"`
	Amount     int                    `json:"amount" db:"amount"`
	Status     string                 `json:"status" db:"status"`
	CreatedAt  time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at" db:"updated_at"`
}

type CreateTaskRequest struct {
	Type       string                 `json:"type" validate:"required"`
	TemplateID string                 `json:"template_id,omitempty"`
	Template   map[string]interface{} `json:"template,omitempty"`
	Amount     int                    `json:"amount" validate:"required,gte=1"`
	Format     string                 `json:"format" validate:"required,oneof=json csv sql"`
}
