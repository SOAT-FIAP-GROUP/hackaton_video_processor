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

func (vr *VideoRepository) GetVideoByID(ctx context.Context, videoID string) (*VideoEntity, error) {
	query := `SELECT id, user_id, name, path, uploaded_at, processed_at FROM videos WHERE id = $1`
	var video VideoEntity
	err := vr.dbConn.QueryRow(ctx, query, func(rows *sql.Rows) error {
		return rows.Scan(&video.ID, &video.UserID, &video.Name, &video.Path, &video.UploadedAt, &video.ProcessedAt)
	}, videoID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Video not found
		}
		return nil, fmt.Errorf("failed to get video by ID: %w", err)
	}

	return &video, nil
}

func (vr *VideoRepository) ListVideosByUserID(ctx context.Context, userID string) ([]*VideoEntity, error) {
	query := `SELECT id, user_id, name, path, uploaded_at, processed_at FROM videos WHERE user_id = $1 ORDER BY uploaded_at DESC`

	var videos []*VideoEntity
	err := vr.dbConn.QueryRows(ctx, query, func(rows *sql.Rows) error {
		var video VideoEntity
		if err := rows.Scan(&video.ID, &video.UserID, &video.Name, &video.Path, &video.UploadedAt, &video.ProcessedAt); err != nil {
			return err
		}
		videos = append(videos, &video)
		return nil
	}, userID)

	if err != nil {
		return nil, fmt.Errorf("failed to list videos by user ID: %w", err)
	}

	return videos, nil
}
