package services

import (
	"context"
	"encoding/json"
	"strconv"
	"template-service/internal/models"
	"template-service/internal/repository"
	"template-service/pkg/db/redis"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type TemplateService interface {
	CreateNewTemplate(ctx context.Context, template models.Template) (int64, error)
	GetTemplateByID(ctx context.Context, id int64) (*models.Template, error)
}

type templateService struct {
	repo   repository.TemplateRepository
	logger *zap.SugaredLogger
	redis  *redis.Redis
}

func NewTemplateService(repo repository.TemplateRepository, redis *redis.Redis, logger *zap.SugaredLogger) *templateService {
	return &templateService{repo: repo, logger: logger, redis: redis}
}

func (t *templateService) CreateNewTemplate(ctx context.Context, template models.Template) (int64, error) {
	template.TemplateID = uuid.New().String()
	template.CreatedAt = time.Now()
	template.UpdatedAt = template.CreatedAt

	id, err := t.repo.CreateNewTemplate(ctx, template)
	if err != nil {
		t.logger.Errorf("Failed to create new template: %v", err)
		return 0, err
	}

	templateData, err := json.Marshal(template)
	if err != nil {
		t.logger.Errorf("Failed to marshal template data: %v", err)
		return id, err
	}

	err = t.redis.Set(ctx, "template:"+strconv.FormatInt(id, 10), templateData, time.Hour)
	if err != nil {
		t.logger.Errorf("Failed to cache template data: %v", err)
	}

	return id, nil
}

func (t *templateService) GetTemplateByID(ctx context.Context, id int64) (*models.Template, error) {
	cacheKey := "template:" + strconv.FormatInt(id, 10)
	templateData, err := t.redis.Get(ctx, cacheKey)
	if err == nil {
		var template models.Template
		if err := json.Unmarshal([]byte(templateData), &template); err == nil {
			t.logger.Debugf("Template %d found in Redis", id)
			return &template, nil
		}
	}

	template, err := t.repo.GetTemplateByID(ctx, id)
	if err != nil {
		t.logger.Errorf("Failed to get template: %v", err)
		return nil, err
	}

	templateDataBytes, err := json.Marshal(template)
	if err != nil {
		t.logger.Warnf("Failed to marshal template %d for caching: %v", id, err)
	} else {
		if err := t.redis.Set(ctx, cacheKey, templateDataBytes, time.Hour); err != nil {
			t.logger.Warnf("Failed to cache template %d: %v", id, err)
		}
	}

	t.logger.Infof("Template retrieved with ID: %d", id)
	return template, nil
}
