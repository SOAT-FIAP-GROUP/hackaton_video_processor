package SQS

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type SQSEmitter struct {
	QueueURL string
	client   *SQSClient
}

func NewSQSEmitter(queueURL string, client *SQSClient) *SQSEmitter {
	return &SQSEmitter{QueueURL: queueURL, client: client}
}

func (e *SQSEmitter) SendMessage(ctx context.Context, message []byte) (string, error) {
	result, err := e.client.client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:     &e.QueueURL,
		MessageBody:  aws.String(string(message)),
		DelaySeconds: 0,
	})
	if err != nil {
		return "", fmt.Errorf("failed to send message: %w", err)
	}

	return *result.MessageId, nil
}
