package kafka

import (
	"context"
	"encoding/json"

	"github.com/dolsom/user-service/internal/model"
	"github.com/dolsom/user-service/internal/repository"
	"github.com/rs/zerolog/log"
	kafka "github.com/segmentio/kafka-go"
)

type Consumer struct {
	broker      string
	profileRepo *repository.ProfileRepository
	activityRepo *repository.ActivityRepository
}

func NewConsumer(
	broker string,
	profileRepo *repository.ProfileRepository,
	activityRepo *repository.ActivityRepository,
) *Consumer {
	return &Consumer{
		broker:       broker,
		profileRepo:  profileRepo,
		activityRepo: activityRepo,
	}
}

func (c *Consumer) Start(ctx context.Context) {
	go c.consume(ctx, "user.registered", c.handleUserRegistered)
	go c.consume(ctx, "subscription.changed", c.handleSubscriptionChanged)
}

func (c *Consumer) consume(ctx context.Context, topic string, handler func([]byte)) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{c.broker},
		Topic:    topic,
		GroupID:  "user-service",
		MinBytes: 1,
		MaxBytes: 1e6,
	})
	defer r.Close()

	log.Info().Str("topic", topic).Msg("kafka consumer started")

	for {
		msg, err := r.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return // context cancelled — clean shutdown
			}
			log.Error().Err(err).Str("topic", topic).Msg("kafka read error")
			continue
		}
		handler(msg.Value)
	}
}

func (c *Consumer) handleUserRegistered(data []byte) {
	var event struct {
		UserID string `json:"user_id"`
		Email  string `json:"email"`
	}
	if err := json.Unmarshal(data, &event); err != nil {
		log.Error().Err(err).Msg("failed to parse user.registered event")
		return
	}

	exists, _ := c.profileRepo.Exists(event.UserID)
	if exists {
		return
	}

	lang := "en"
	tz := "UTC"
	if err := c.profileRepo.Create(&model.Profile{
		UserID:   event.UserID,
		Email:    event.Email,
		Language: lang,
		Timezone: tz,
	}); err != nil {
		log.Error().Err(err).Str("user_id", event.UserID).Msg("failed to create profile")
		return
	}

	log.Info().Str("user_id", event.UserID).Msg("profile created from user.registered event")
}

func (c *Consumer) handleSubscriptionChanged(data []byte) {
	var event struct {
		UserID string `json:"user_id"`
		Status string `json:"status"`
		Plan   string `json:"plan"`
	}
	if err := json.Unmarshal(data, &event); err != nil {
		log.Error().Err(err).Msg("failed to parse subscription.changed event")
		return
	}

	meta, _ := json.Marshal(event)
	if err := c.activityRepo.Log(event.UserID, model.ActivitySubscriptionChanged, string(meta)); err != nil {
		log.Error().Err(err).Str("user_id", event.UserID).Msg("failed to log subscription activity")
	}
}
