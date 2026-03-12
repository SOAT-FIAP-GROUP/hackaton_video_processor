package main

import (
	"context"
	"database/sql"
	"fmt"
	delivery "frontend/internal/api/http"
	"frontend/internal/usecase"
	"log"
	"shared/SQS"
	"shared/config"
	"shared/database/connection"
	"shared/database/connection/migrations"
	"shared/database/repository"
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

	err = c.ValidateDBConfig(true)
	if err != nil {
		log.Fatalf("Invalid configuration for Database: %v", err)
	}

	conn, err := connection.CreatePostgresConnection(c.DBHost, c.DBUser, c.DBPassword, c.DBReplicaName, c.DBSSLMode, c.DBPort)
	if err != nil {
		panic("Failed to create database connection: " + err.Error())
	}
	defer conn.Close()

	connVar := migrations.CreateConnectionVar(c.DBHost, c.DBName, c.DBReplicaName, c.DBUser, c.DBPassword, c.DBPort)
	err = conn.QueryRow(context.Background(), connVar, func(rows *sql.Rows) error {
		var result int
		if rows.Next() {
			err := rows.Scan(&result)
			if err != nil {
				return err
			}
			log.Printf("Read instance test query result: %d", result)
		}
		return nil
	})
	if err != nil {
		log.Fatalf("Failed to generate db var: %s", err.Error())
	}

	m := migrations.GetMigrationsForReadInstance()
	err = conn.RunMigrations(m)
	if err != nil {
		log.Fatalf("Failed to run migrations: %s", err.Error())
	}

	r := repository.NewVideoRepository(conn)

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

	videoHandler := delivery.NewVideoHandler(s3, emitter, r)
	userHandler := delivery.NewUserHandler(cognitoUseCase)

	router := delivery.SetupRouter(videoHandler, userHandler, cognitoUseCase)

	fmt.Println("🎬 Servidor iniciado na porta", c.APIPort)
	fmt.Println(fmt.Sprintf("📂 Acesse: http://localhost:%v", c.APIPort))

	addr := fmt.Sprintf(":%v", c.APIPort)

	if err = router.Engine.Run(addr); err != nil {
		log.Fatalf("Erro ao iniciar servidor: %v", err)
	}
}
