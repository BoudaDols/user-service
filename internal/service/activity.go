package service

import (
	"github.com/dolsom/user-service/internal/model"
	"github.com/dolsom/user-service/internal/repository"
)

type ActivityService struct {
	repo *repository.ActivityRepository
}

func NewActivityService(repo *repository.ActivityRepository) *ActivityService {
	return &ActivityService{repo: repo}
}

func (s *ActivityService) GetByUserID(userID string, limit, offset int) ([]model.ActivityLog, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.GetByUserID(userID, limit, offset)
}
