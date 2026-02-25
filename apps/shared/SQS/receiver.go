package SQS

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type SQSReceiver struct {
	QueueURL string
	client   *SQSClient
}

func NewSQSReceiver(queueURL string, client *SQSClient) *SQSReceiver {
	return &SQSReceiver{QueueURL: queueURL, client: client}
}

func (l *SQSReceiver) Receive(ctx context.Context) (SQSMessage, error) {
	result, err := l.client.client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(l.QueueURL),
		MaxNumberOfMessages: 1,
		WaitTimeSeconds:     10,
	})
	if err != nil {
		return SQSMessage{}, fmt.Errorf("error trying to receive a message: %w", err)
	}

	if len(result.Messages) == 0 {
		return SQSMessage{}, nil
	}

	msg := result.Messages[0]

	return SQSMessage{
		MessageId:     *msg.MessageId,
		ReceiptHandle: *msg.ReceiptHandle,
		Content:       []byte(*msg.Body),
	}, nil
}

func (l *SQSReceiver) AckMessage(ctx context.Context, messageId *string) error {
	_, err := l.client.client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(l.QueueURL),
		ReceiptHandle: messageId,
	})
	if err != nil {
		return fmt.Errorf("error trying to delete a message: %w", err)
	}

	return nil
}
