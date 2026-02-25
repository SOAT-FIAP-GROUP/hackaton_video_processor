package SQS

import "time"

type SQSMessage struct {
	MessageId     string
	ReceiptHandle string
	Content       []byte
}

type BrokerMessage struct {
	VideoPath string    `json:"video_path"`
	UserID    string    `json:"user_id"`
	UploadAt  time.Time `json:"upload_at"`
}
