package service

import (
	"encoding/json"

	"github.com/dolsom/user-service/internal/kafka"
	"github.com/dolsom/user-service/internal/model"
	"github.com/dolsom/user-service/internal/repository"
)

type PreferenceService struct {
	repo     *repository.PreferenceRepository
	activity *repository.ActivityRepository
	producer *kafka.Producer
}

func NewPreferenceService(
	repo *repository.PreferenceRepository,
	activity *repository.ActivityRepository,
	producer *kafka.Producer,
) *PreferenceService {
	return &PreferenceService{repo: repo, activity: activity, producer: producer}
}

func (s *PreferenceService) GetAll(userID string) ([]model.Preference, error) {
	return s.repo.GetAll(userID)
}

func (s *PreferenceService) Upsert(userID, key, value string) error {
	if err := s.repo.Upsert(userID, key, value); err != nil {
		return err
	}

	meta, _ := json.Marshal(map[string]string{"key": key, "value": value})
	_ = s.activity.Log(userID, model.ActivityPreferencesUpdated, string(meta))

	_ = s.producer.Publish("user.preferences_updated", map[string]any{
		"user_id": userID,
		"key":     key,
		"value":   value,
	})

	return nil
}

func (s *PreferenceService) Delete(userID, key string) error {
	return s.repo.Delete(userID, key)
}
