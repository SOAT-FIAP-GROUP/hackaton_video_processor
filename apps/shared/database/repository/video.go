package repository

import (
	"context"
	"database/sql"
	"fmt"
	"shared/database/connection"
)

type VideoRepository struct {
	dbConn connection.DatabaseConnection
}

func NewVideoRepository(dbConn connection.DatabaseConnection) *VideoRepository {
	return &VideoRepository{
		dbConn: dbConn,
	}
}

func (vr *VideoRepository) CreateVideo(ctx context.Context, video *VideoEntity) error {
	query := `INSERT INTO videos (user_id, name, path, uploaded_at) VALUES ($1, $2, $3, $4) RETURNING id, processed_at`
	err := vr.dbConn.QueryRow(ctx, query, func(rows *sql.Rows) error {
		return rows.Scan(&video.ID, &video.ProcessedAt)
	}, video.UserID, video.Name, video.Path, video.UploadedAt)
	if err != nil {
		return fmt.Errorf("failed to create video: %w", err)
	}

	return nil
}
