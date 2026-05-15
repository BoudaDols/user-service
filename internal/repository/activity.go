package repository

import (
	"database/sql"

	"github.com/dolsom/user-service/internal/model"
)

type ActivityRepository struct {
	db *sql.DB
}

func NewActivityRepository(db *sql.DB) *ActivityRepository {
	return &ActivityRepository{db: db}
}

func (r *ActivityRepository) Log(userID, activityType, metadata string) error {
	_, err := r.db.Exec(
		`INSERT INTO activity_logs (user_id, type, metadata) VALUES (?, ?, ?)`,
		userID, activityType, metadata,
	)
	return err
}

func (r *ActivityRepository) GetByUserID(userID string, limit, offset int) ([]model.ActivityLog, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, type, metadata, created_at
		 FROM activity_logs WHERE user_id = ?
		 ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []model.ActivityLog
	for rows.Next() {
		var l model.ActivityLog
		if err := rows.Scan(&l.ID, &l.UserID, &l.Type, &l.Metadata, &l.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, rows.Err()
}
