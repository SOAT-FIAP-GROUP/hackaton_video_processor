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
	"sync"
	"time"
)

type Setup struct {
	c       *SQS.SQSClient
	r       *SQS.SQSReceiver
	h       *handlers.MessageHandler
	workers int
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
		c:       client,
		r:       receiver,
		h:       handler,
		workers: c.NumberWorkers,
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
	msgCh := make(chan SQS.SQSMessage, s.workers)

	procCtx, procCancel := context.WithCancel(context.Background())
	defer procCancel()

	var wg sync.WaitGroup
	for range s.workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for msg := range msgCh {
				if err := s.h.HandleMessage(procCtx, msg.Content); err != nil {
					log.Println("Failed to handle message:", err)
					continue
				}

				if err := s.r.AckMessage(procCtx, &msg.ReceiptHandle); err != nil {
					log.Println("Failed to ack message:", err)
					continue
				}

				log.Println("Message acknowledged! ID:", msg.MessageId)
			}
		}()
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("Shutting down worker, draining in-flight messages...")
			close(msgCh)
			wg.Wait()
			procCancel()
			log.Println("Worker shutdown complete.")
			return nil

		default:
			msg, err := s.r.Receive(ctx)
			if err != nil {
				close(msgCh)
				wg.Wait()
				return fmt.Errorf("failed to receive message: %v", err)
			}

			if msg.MessageId == "" {
				log.Println("No messages received, waiting...")
				time.Sleep(5 * time.Second)
				continue
			}

			msgCh <- msg
		}
	}
}
