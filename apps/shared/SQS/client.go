package SQS

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type SQSClient struct {
	client *sqs.Client
}

func NewSQSClient(region string) (*SQSClient, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	client := sqs.NewFromConfig(cfg)

	return &SQSClient{client: client}, nil
}
