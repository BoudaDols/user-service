package model

import "time"

type Profile struct {
	ID          int64     `json:"-"`
	UserID      string    `json:"user_id"`
	Email       string    `json:"email"`
	DisplayName *string   `json:"display_name"`
	AvatarURL   *string   `json:"avatar_url"`
	Bio         *string   `json:"bio"`
	Language    string    `json:"language"`
	Timezone    string    `json:"timezone"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Preference struct {
	ID        int64     `json:"-"`
	UserID    string    `json:"user_id"`
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ActivityLog struct {
	ID        int64     `json:"id"`
	UserID    string    `json:"user_id"`
	Type      string    `json:"type"`
	Metadata  string    `json:"metadata"`
	CreatedAt time.Time `json:"created_at"`
}

const (
	ActivityProfileUpdated      = "profile_updated"
	ActivityPreferencesUpdated  = "preferences_updated"
	ActivitySubscriptionChanged = "subscription_changed"
	ActivityAPIRequest          = "api_request"
)
