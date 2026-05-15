package repository

import (
	"database/sql"

	"github.com/dolsom/user-service/internal/model"
)

type PreferenceRepository struct {
	db *sql.DB
}

func NewPreferenceRepository(db *sql.DB) *PreferenceRepository {
	return &PreferenceRepository{db: db}
}

func (r *PreferenceRepository) GetAll(userID string) ([]model.Preference, error) {
	rows, err := r.db.Query(
		`SELECT user_id, key_name, value, created_at, updated_at FROM preferences WHERE user_id = ?`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prefs []model.Preference
	for rows.Next() {
		var p model.Preference
		if err := rows.Scan(&p.UserID, &p.Key, &p.Value, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		prefs = append(prefs, p)
	}
	return prefs, rows.Err()
}

func (r *PreferenceRepository) Upsert(userID, key, value string) error {
	_, err := r.db.Exec(
		`INSERT INTO preferences (user_id, key_name, value)
		 VALUES (?, ?, ?)
		 ON DUPLICATE KEY UPDATE value = ?, updated_at = NOW()`,
		userID, key, value, value,
	)
	return err
}

func (r *PreferenceRepository) Delete(userID, key string) error {
	_, err := r.db.Exec(
		`DELETE FROM preferences WHERE user_id = ? AND key_name = ?`,
		userID, key,
	)
	return err
}
