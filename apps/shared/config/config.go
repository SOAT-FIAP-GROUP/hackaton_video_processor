package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	CognitoClientID     string `env:"COGNITO_CLIENT_ID,required"`
	CognitoClientSecret string `env:"COGNITO_CLIENT_SECRET,required"`
	CognitoUserPoolID   string `env:"COGNITO_USER_POOL_ID,required"`
	AWSRegion           string `env:"AWS_REGION,required"`
	AWSAccessKeyID      string `env:"AWS_ACCESS_KEY_ID,required"`
	AWSSecretAccessKey  string `env:"AWS_SECRET_ACCESS_KEY,required"`
	AWSS3Bucket         string `env:"AWS_S3_BUCKET,required"`
	AWSSQSQueueName     string `env:"AWS_SQS_QUEUE_NAME,required"`
	TempPath            string `env:"TEMP_PATH,required"`
	DBHost              string `env:"DB_HOST,required"`
	DBPort              int    `env:"DB_PORT,required"`
	DBUser              string `env:"DB_USER,required"`
	DBPassword          string `env:"DB_PASSWORD,required"`
	DBName              string `env:"DB_NAME,required"`
	DBSSLMode           string `env:"DB_SSL_MODE,required"`
	NumberWorkers       int    `env:"NUMBER_WORKERS,required"`
}

func NewConfig() *Config {
	dbPort, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		dbPort = 5432
	}

	workers, err := strconv.Atoi(os.Getenv("NUMBER_WORKERS"))
	if err != nil {
		workers = 5
	}

	return &Config{
		CognitoClientSecret: os.Getenv("COGNITO_CLIENT_SECRET"),
		CognitoClientID:     os.Getenv("COGNITO_CLIENT_ID"),
		CognitoUserPoolID:   os.Getenv("COGNITO_USER_POOL_ID"),
		AWSRegion:           os.Getenv("AWS_REGION"),
		AWSAccessKeyID:      os.Getenv("AWS_ACCESS_KEY_ID"),
		AWSSecretAccessKey:  os.Getenv("AWS_SECRET_ACCESS_KEY"),
		AWSS3Bucket:         os.Getenv("AWS_S3_BUCKET"),
		AWSSQSQueueName:     os.Getenv("AWS_SQS_QUEUE_NAME"),
		TempPath:            os.Getenv("TEMP_PATH"),
		DBHost:              os.Getenv("DB_HOST"),
		DBPort:              dbPort,
		DBUser:              os.Getenv("DB_USER"),
		DBPassword:          os.Getenv("DB_PASSWORD"),
		DBName:              os.Getenv("DB_NAME"),
		DBSSLMode:           os.Getenv("DB_SSL_MODE"),
		NumberWorkers:       workers,
	}
}

func (c *Config) ValidateCognitoConfig() error {
	if c.CognitoClientID == "" {
		return fmt.Errorf("cognito client ID is required")
	}

	if c.CognitoClientSecret == "" {
		return fmt.Errorf("cognito client secret is required")
	}

	if c.CognitoUserPoolID == "" {
		return fmt.Errorf("cognito user pool ID is required")
	}

	if c.AWSRegion == "" {
		return fmt.Errorf("AWS region is required")
	}

	if c.AWSAccessKeyID == "" {
		return fmt.Errorf("AWS access key ID is required")
	}

	if c.AWSSecretAccessKey == "" {
		return fmt.Errorf("AWS secret access key is required")
	}

	return nil
}

func (c *Config) ValidateS3Config() error {
	if c.AWSRegion == "" {
		return fmt.Errorf("AWS region is required")
	}

	if c.AWSAccessKeyID == "" {
		return fmt.Errorf("AWS access key ID is required")
	}

	if c.AWSSecretAccessKey == "" {
		return fmt.Errorf("AWS secret access key is required")
	}

	if c.AWSS3Bucket == "" {
		return fmt.Errorf("AWS S3 bucket name is required")
	}

	if c.TempPath == "" {
		return fmt.Errorf("Temp path is required")
	}

	return nil
}

func (c *Config) ValidateSQSConfig() error {
	if c.AWSRegion == "" {
		return fmt.Errorf("AWS region is required")
	}

	if c.AWSAccessKeyID == "" {
		return fmt.Errorf("AWS access key ID is required")
	}

	if c.AWSSecretAccessKey == "" {
		return fmt.Errorf("AWS secret access key is required")
	}

	if c.AWSSQSQueueName == "" {
		return fmt.Errorf("AWS SQS queue name is required")
	}

	return nil
}

func (c *Config) ValidateDBConfig() error {
	if c.DBHost == "" {
		return fmt.Errorf("DB host is required")
	}

	if c.DBUser == "" {
		return fmt.Errorf("DB user is required")
	}

	if c.DBPassword == "" {
		return fmt.Errorf("DB password is required")
	}

	if c.DBName == "" {
		return fmt.Errorf("DB name is required")
	}

	if c.DBSSLMode == "" {
		return fmt.Errorf("DB SSL mode is required")
	}

	return nil
}
