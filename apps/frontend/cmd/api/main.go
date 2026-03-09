package main

import (
	"context"
	"fmt"
	delivery "frontend/internal/api/http"
	"frontend/internal/usecase"
	"log"
	"shared/SQS"
	"shared/config"
	"shared/storage/S3"
)

func main() {
	c := config.NewConfig()
	err := c.ValidateCognitoConfig()
	if err != nil {
		log.Fatalf("Invalid configuration for Cognito: %v", err)
	}

	err = c.ValidateS3Config()
	if err != nil {
		log.Fatalf("Invalid configuration for S3: %v", err)
	}

	err = c.ValidateSQSConfig()
	if err != nil {
		log.Fatalf("Invalid configuration for SQS: %v", err)
	}

	cognitoUseCase, err := usecase.NewCognitoClient(c.AWSRegion, c.CognitoClientID, c.CognitoClientSecret, c.CognitoUserPoolID)
	if err != nil {
		log.Fatalf("Failed to initialize Cognito client: %v", err)
	}

	s3, err := S3.NewS3Client(context.Background(), c.AWSS3Bucket, c.AWSRegion, c.TempPath)
	if err != nil {
		log.Fatalf("Failed to initialize S3 client: %v", err)
	}

	sqs, err := SQS.NewSQSClient(c.AWSRegion)
	if err != nil {
		log.Fatalf("Failed to initialize SQS client: %v", err)
	}

	emitter := SQS.NewSQSEmitter(c.AWSSQSQueueNameVideoProcessing, sqs)

	videoHandler := delivery.NewVideoHandler(s3, emitter)
	userHandler := delivery.NewUserHandler(cognitoUseCase)

	router := delivery.SetupRouter(videoHandler, userHandler, cognitoUseCase)

	fmt.Println("🎬 Servidor iniciado na porta 8080")
	fmt.Println("📂 Acesse: http://localhost:8080")

	if err = router.Engine.Run(":8080"); err != nil {
		log.Fatalf("Erro ao iniciar servidor: %v", err)
	}
}
