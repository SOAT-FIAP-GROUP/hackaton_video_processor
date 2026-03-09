package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"processor/internal/usecase"
	"shared/SQS"
	"shared/database/repository"
	"shared/storage/S3"
	"strings"
)

type MessageHandler struct {
	s3   *S3.S3Client
	p    *usecase.VideoProcessingUseCase
	repo *repository.VideoRepository
}

func NewMessageHandler(s3 *S3.S3Client, processor *usecase.VideoProcessingUseCase, repo *repository.VideoRepository) *MessageHandler {
	return &MessageHandler{
		s3:   s3,
		p:    processor,
		repo: repo,
	}
}

func (h *MessageHandler) HandleMessage(ctx context.Context, message []byte) error {
	var m SQS.BrokerMessage

	err := json.Unmarshal(message, &m)
	if err != nil {
		return fmt.Errorf("error unmarshalling message: %s", err)
	}

	keyProcess := m.VideoPath

	log.Println("Message received: ", m)

	key, err := h.s3.DownloadToTempDir(ctx, m.VideoPath)
	if err != nil {
		return fmt.Errorf("error downloading video: %s", err)
	}

	log.Println("Downloaded from S3 for video. Process: ", keyProcess)

	framesPath, zipPath, err := h.p.Process(key)
	if err != nil {
		return fmt.Errorf("error processing video: %s", err)
	}

	log.Println("Processed video. Process: ", keyProcess)

	zipFileNameSlice := strings.Split(zipPath, "/")
	zipFileName := zipFileNameSlice[len(zipFileNameSlice)-1]

	s3FileName := fmt.Sprintf("%s/%s/%s", "downloads", m.UserID, zipFileName)

	err = h.s3.UploadFile(ctx, zipPath, s3FileName)
	if err != nil {
		return fmt.Errorf("error uploading file: %s", err)
	}

	log.Println("Uploaded file. Process: ", keyProcess)

	err = h.p.DeleteLocalFiles(zipPath, framesPath, key)
	if err != nil {
		log.Printf("error deleting files: %s", err)
	}

	log.Println("Deleted local files. Process: ", keyProcess)

	entity := &repository.VideoEntity{
		UserID:     m.UserID,
		Name:       m.UserName,
		Path:       s3FileName,
		UploadedAt: m.UploadAt,
	}

	err = h.repo.CreateVideo(ctx, entity)
	if err != nil {
		log.Printf("error saving video to database: %s", err)
	}

	log.Printf("Downloaded video: %s", key)

	return nil
}
