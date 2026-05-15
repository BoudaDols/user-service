package service

import (
	"encoding/json"
	"errors"

	"github.com/dolsom/user-service/internal/kafka"
	"github.com/dolsom/user-service/internal/model"
	"github.com/dolsom/user-service/internal/repository"
)

var ErrProfileNotFound = errors.New("profile not found")
var ErrProfileAlreadyExists = errors.New("profile already exists")

type ProfileService struct {
	repo     *repository.ProfileRepository
	activity *repository.ActivityRepository
	producer *kafka.Producer
}

func NewProfileService(
	repo *repository.ProfileRepository,
	activity *repository.ActivityRepository,
	producer *kafka.Producer,
) *ProfileService {
	return &ProfileService{repo: repo, activity: activity, producer: producer}
}

func (s *ProfileService) Create(userID, email string) (*model.Profile, error) {
	exists, err := s.repo.Exists(userID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrProfileAlreadyExists
	}

	lang := "en"
	tz := "UTC"
	p := &model.Profile{
		UserID:   userID,
		Email:    email,
		Language: lang,
		Timezone: tz,
	}

	if err := s.repo.Create(p); err != nil {
		return nil, err
	}

	return s.repo.GetByUserID(userID)
}

func (s *ProfileService) Get(userID string) (*model.Profile, error) {
	p, err := s.repo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, ErrProfileNotFound
	}
	return p, nil
}

func (s *ProfileService) Update(userID string, fields map[string]any) (*model.Profile, error) {
	exists, err := s.repo.Exists(userID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrProfileNotFound
	}

	if err := s.repo.Update(userID, fields); err != nil {
		return nil, err
	}

	p, err := s.repo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}

	// Log activity
	meta, _ := json.Marshal(fields)
	_ = s.activity.Log(userID, model.ActivityProfileUpdated, string(meta))

	// Publish Kafka event
	_ = s.producer.Publish("user.profile_updated", map[string]any{
		"user_id": userID,
		"fields":  fields,
	})

	return p, nil
}
