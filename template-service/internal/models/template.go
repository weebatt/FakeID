package models

import "time"

type Template struct {
	ID         int64                  `json:"id" db:"id"`
	TemplateID string                 `json:"template_id" db:"template_id"`
	UserID     string                 `json:"user_id" db:"user_id"`
	Title      string                 `json:"title" db:"title"`
	Content    map[string]interface{} `json:"content" db:"content"`
	CreatedAt  time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at" db:"updated_at"`
}

type CreateTemplateRequest struct {
	Title   string                 `json:"title" validate:"required"`
	Content map[string]interface{} `json:"content" validate:"required"`
}
