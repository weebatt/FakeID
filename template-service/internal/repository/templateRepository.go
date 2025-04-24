package repository

import (
	"context"
	"template-service/internal/models"
	"template-service/pkg/db/postgres"

	"go.uber.org/zap"
)

type TemplateRepository interface {
	GetTemplateByID(ctx context.Context, id int64) (*models.Template, error)
	CreateNewTemplate(ctx context.Context, template models.Template) (int64, error)
}

type templateRepository struct {
	db     *postgres.DB
	logger *zap.SugaredLogger
}

func NewTemplateRepository(pool *postgres.DB, logger *zap.SugaredLogger) *templateRepository {
	return &templateRepository{db: pool, logger: logger}
}

func (t *templateRepository) CreateNewTemplate(ctx context.Context, template models.Template) (int64, error) {
	query := `
		INSERT INTO templates (template_id, user_id, title, content, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
		`
	var id int64
	err := t.db.QueryRow(ctx, query, template.TemplateID, template.UserID, template.Title, template.Content, template.CreatedAt, template.UpdatedAt).Scan(&id)
	if err != nil {
		t.logger.Errorf("Failed to insert template: %v", err)
		return 0, err
	}

	return id, nil
}

func (t *templateRepository) GetTemplateByID(ctx context.Context, id int64) (*models.Template, error) {
	query := `
		SELECT id, template_id, user_id, title, content, created_at, updated_at
		FROM templates
		WHERE id = $1
		`

	var template models.Template
	err := t.db.QueryRow(ctx, query, id).Scan(&template.ID, &template.TemplateID, &template.UserID, &template.Title, &template.Content, &template.CreatedAt, &template.UpdatedAt)
	if err != nil {
		t.logger.Errorf("Failed to get template: %v", err)
		return nil, err
	}

	return &template, nil
}
