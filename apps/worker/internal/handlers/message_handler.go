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
	"time"
)

type MessageHandler struct {
	s3      *S3.S3Client
	p       *usecase.VideoProcessingUseCase
	repo    *repository.VideoRepository
	emitter *SQS.SQSEmitter
}

func NewMessageHandler(s3 *S3.S3Client, processor *usecase.VideoProcessingUseCase, repo *repository.VideoRepository, emitter *SQS.SQSEmitter) *MessageHandler {
	return &MessageHandler{
		s3:      s3,
		p:       processor,
		repo:    repo,
		emitter: emitter,
	}
}

func (h *MessageHandler) HandleMessage(ctx context.Context, message []byte) error {
	ctx = context.WithoutCancel(ctx)
	var m SQS.BrokerMessage

	err := json.Unmarshal(message, &m)
	if err != nil {
		notification := SQS.NotificationMessage{
			ProcessingID:     m.UserID,
			UserName:         m.UserName,
			UserEmail:        m.UserEmail,
			UserID:           m.UserID,
			NotificationDate: time.Now(),
		}

		h.sendNotification(ctx, notification)
		return fmt.Errorf("error unmarshalling message: %s", err)
	}

	keyProcess := m.VideoPath

	log.Println("Message received: ", m)

	key, err := h.s3.DownloadToTempDir(ctx, m.VideoPath)
	if err != nil {
		notification := SQS.NotificationMessage{
			ProcessingID:     m.UserID,
			UserName:         m.UserName,
			UserEmail:        m.UserEmail,
			UserID:           m.UserID,
			NotificationDate: time.Now(),
		}

		h.sendNotification(ctx, notification)
		return fmt.Errorf("error downloading video: %s", err)
	}

	log.Println("Downloaded from S3 for video. Process: ", keyProcess)

	framesPath, zipPath, err := h.p.Process(ctx, key)
	if err != nil {
		notification := SQS.NotificationMessage{
			ProcessingID:     m.UserID,
			UserName:         m.UserName,
			UserEmail:        m.UserEmail,
			UserID:           m.UserID,
			NotificationDate: time.Now(),
		}

		h.sendNotification(ctx, notification)
		return fmt.Errorf("error processing video: %s", err)
	}

	log.Println("Processed video. Process: ", keyProcess)

	zipFileNameSlice := strings.Split(zipPath, "/")
	zipFileName := zipFileNameSlice[len(zipFileNameSlice)-1]

	s3FileName := fmt.Sprintf("%s/%s/%s", "downloads", m.UserID, zipFileName)

	err = h.s3.UploadFile(ctx, zipPath, s3FileName)
	if err != nil {
		notification := SQS.NotificationMessage{
			ProcessingID:     m.UserID,
			UserName:         m.UserName,
			UserEmail:        m.UserEmail,
			UserID:           m.UserID,
			NotificationDate: time.Now(),
		}

		h.sendNotification(ctx, notification)
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
		Name:       m.FileName,
		Path:       s3FileName,
		UploadedAt: m.UploadAt,
	}

	err = h.repo.CreateVideo(ctx, entity)
	if err != nil {
		notification := SQS.NotificationMessage{
			ProcessingID:     m.UserID,
			UserName:         m.UserName,
			UserEmail:        m.UserEmail,
			UserID:           m.UserID,
			NotificationDate: time.Now(),
		}

		h.sendNotification(ctx, notification)

		log.Printf("error saving video to database: %s", err)
	}

	log.Printf("Downloaded video: %s", key)

	/*notification := SQS.NotificationMessage{
		ProcessingID:     entity.UserID,
		UserName:         m.UserName,
		UserEmail:        m.UserEmail,
		UserID:           m.UserID,
		NotificationDate: time.Now(),
	}

	jsonNotification, err := json.Marshal(notification)
	if err != nil {
		log.Printf("error marshalling notification message: %s", err)
	}

	r, err := h.emitter.SendMessage(ctx, jsonNotification)
	if err != nil {
		log.Printf("error sending notification message: %s", err)
	}

	log.Printf("Notification sent! Message ID: %s", r)*/

	return nil
}

func (h *MessageHandler) sendNotification(ctx context.Context, notification SQS.NotificationMessage) {
	jsonNotification, err := json.Marshal(notification)
	if err != nil {
		log.Printf("error marshalling notification message: %s", err)
	}

	r, err := h.emitter.SendMessage(ctx, jsonNotification)
	if err != nil {
		log.Printf("error sending notification message: %s", err)
	}

	log.Printf("Notification sent! Message ID: %s", r)
}
