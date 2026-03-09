package internal

import (
	"context"
	"fmt"
	"log"
	"os"
	"processor/internal/handlers"
	"processor/internal/usecase"
	"shared/SQS"
	"shared/config"
	"shared/database/connection"
	"shared/database/repository"
	"shared/storage/S3"
	"time"
)

type Setup struct {
	c *SQS.SQSClient
	r *SQS.SQSReceiver
	h *handlers.MessageHandler
}

func NewSetup() (*Setup, error) {
	c := config.NewConfig()

	err := c.ValidateS3Config()
	if err != nil {
		return nil, fmt.Errorf("s3 Config error: %v", err)
	}

	err = c.ValidateSQSConfig()
	if err != nil {
		return nil, fmt.Errorf("SQS Config error: %v", err)
	}

	err = c.ValidateDBConfig()
	if err != nil {
		return nil, fmt.Errorf("DB Config error: %v", err)
	}

	dbConn, err := connection.CreatePostgresConnection(c.DBHost, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode, c.DBPort)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to database: %v", err)
	}

	videoRepository := repository.NewVideoRepository(dbConn)

	client, err := SQS.NewSQSClient(c.AWSRegion)
	if err != nil {
		return nil, fmt.Errorf("Failed to create SQS client: %v", err)
	}

	receiver := SQS.NewSQSReceiver(c.AWSSQSQueueName, client)

	ctx := context.Background()

	s3, err := S3.NewS3Client(ctx, c.AWSS3Bucket, c.AWSRegion, c.TempPath+"/downloads")
	if err != nil {
		return nil, fmt.Errorf("Failed to create S3 client: %v", err)
	}

	vpuc := usecase.NewVideoProcessingUseCase(c.TempPath+"/frames", c.TempPath+"/zips")

	handler := handlers.NewMessageHandler(s3, vpuc, videoRepository)

	err = createTempDirs(c.TempPath)
	if err != nil {
		return nil, fmt.Errorf("Failed to create temp dirs: %v", err)
	}

	return &Setup{
		c: client,
		r: receiver,
		h: handler,
	}, nil
}

func createTempDirs(tempPath string) error {
	if err := os.MkdirAll(tempPath+"/downloads", 0755); err != nil {
		return fmt.Errorf("failed to create temp dir for downloads: %w", err)
	}

	if err := os.MkdirAll(tempPath+"/frames", 0755); err != nil {
		return fmt.Errorf("failed to create temp dir for zips: %w", err)
	}

	if err := os.MkdirAll(tempPath+"/zips", 0755); err != nil {
		return fmt.Errorf("failed to create temp dir for zips: %w", err)
	}

	return nil
}

func (s *Setup) RunWorker(ctx context.Context) error {
	for {
		msg, err := s.r.Receive(ctx)
		if err != nil {
			return fmt.Errorf("failed to receive message: %v", err)
		}

		if msg.MessageId == "" {
			log.Println("No messages received, waiting...")
			time.Sleep(5 * time.Second)
			continue
		}

		err = s.h.HandleMessage(ctx, msg.Content)
		if err != nil {
			log.Println("Failed to handle message:", err)
			return fmt.Errorf("failed to handle message: %v", err)
		}

		err = s.r.AckMessage(ctx, &msg.ReceiptHandle)
		if err != nil {
			return fmt.Errorf("failed to ack message: %v", err)
		}

		log.Println("Message acknowledged! ID:", msg.MessageId)
	}
}
