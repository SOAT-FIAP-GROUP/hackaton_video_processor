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
	UserName  string    `json:"user_name"`
	UserEmail string    `json:"user_email"`
	UploadAt  time.Time `json:"upload_at"`
	FileName  string    `json:"file_name"`
}

type NotificationMessage struct {
	ProcessingID     string    `json:"processamentoId"`
	UserName         string    `json:"username"`
	UserEmail        string    `json:"email"`
	UserID           string    `json:"userId"`
	NotificationDate time.Time `json:"notification_date"`
}
