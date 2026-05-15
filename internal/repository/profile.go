package repository

import (
	"database/sql"
	"fmt"

	"github.com/dolsom/user-service/internal/model"
)

type ProfileRepository struct {
	db *sql.DB
}

func NewProfileRepository(db *sql.DB) *ProfileRepository {
	return &ProfileRepository{db: db}
}

func (r *ProfileRepository) Create(p *model.Profile) error {
	_, err := r.db.Exec(
		`INSERT INTO profiles (user_id, email, display_name, avatar_url, bio, language, timezone)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		p.UserID, p.Email, p.DisplayName, p.AvatarURL, p.Bio, p.Language, p.Timezone,
	)
	return err
}

func (r *ProfileRepository) GetByUserID(userID string) (*model.Profile, error) {
	p := &model.Profile{}
	err := r.db.QueryRow(
		`SELECT user_id, email, display_name, avatar_url, bio, language, timezone, created_at, updated_at
		 FROM profiles WHERE user_id = ?`, userID,
	).Scan(&p.UserID, &p.Email, &p.DisplayName, &p.AvatarURL, &p.Bio, &p.Language, &p.Timezone, &p.CreatedAt, &p.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	return p, err
}

func (r *ProfileRepository) Update(userID string, fields map[string]any) error {
	if len(fields) == 0 {
		return nil
	}

	query := "UPDATE profiles SET updated_at = NOW()"
	args := []any{}

	for col, val := range fields {
		query += fmt.Sprintf(", %s = ?", col)
		args = append(args, val)
	}

	query += " WHERE user_id = ?"
	args = append(args, userID)

	_, err := r.db.Exec(query, args...)
	return err
}

func (r *ProfileRepository) Exists(userID string) (bool, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM profiles WHERE user_id = ?`, userID).Scan(&count)
	return count > 0, err
}
