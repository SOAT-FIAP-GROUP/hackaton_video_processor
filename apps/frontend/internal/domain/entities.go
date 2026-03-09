package domain

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Video struct {
	ID            string
	Filename      string
	Path          string
	UploadedAt    time.Time
	UserId        string
	Processed     bool
	ProcessedPath string
	FileSizeBytes int64
}

type FileInfo struct {
	Filename    string
	Size        int64
	CreatedAt   time.Time
	DownloadURL string
}

type User struct {
	Id       string
	Name     string
	Email    string
	Password string
}

type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	jwt.RegisteredClaims
}
