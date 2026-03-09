package repository

import "time"

type VideoEntity struct {
	ID          int64     `db:"id"`
	UserID      string    `db:"user_id"`
	Name        string    `db:"name"`
	Path        string    `db:"path"`
	UploadedAt  time.Time `db:"uploaded_at"`
	ProcessedAt time.Time `db:"processed_at"`
}
