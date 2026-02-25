package main

import (
	"context"
	"shared/SQS"
)

func main() {
	/*cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err != nil {
		log.Fatal("failed to load config:", err)
	}

	client := sqs.NewFromConfig(cfg)

	queueURL := "https://sqs.us-east-1.amazonaws.com/027419662160/test-queue"

	result, err := client.SendMessage(context.TODO(), &sqs.SendMessageInput{
		QueueUrl:    aws.String(queueURL),
		MessageBody: aws.String(`{"event": "user_signed_up", "userId": "43"}`),
		// Optional: delay in seconds (0-900)
		DelaySeconds: 0,
	})
	if err != nil {
		log.Fatal("failed to send message:", err)
	}

	fmt.Println("Message sent! ID:", *result.MessageId)*/

	client, err := SQS.NewSQSClient("us-east-1")
	if err != nil {
		panic(err)
	}

	emitter := SQS.NewSQSEmitter("https://sqs.us-east-1.amazonaws.com/027419662160/test-queue", client)

	messageId, err := emitter.SendMessage(context.TODO(), []byte(`{"event": "user_signed_up", "userId": "43"}`))
	if err != nil {
		panic(err)
	}

	println("Message sent! ID:", messageId)
}
