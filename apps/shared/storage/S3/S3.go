package S3

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Client struct {
	client        *s3.Client
	presignClient *s3.PresignClient
	bucket        string
	tempPath      string
}

func NewS3Client(ctx context.Context, bucket, region, tempPath string) (*S3Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	client := s3.NewFromConfig(cfg)
	presignClient := s3.NewPresignClient(client)

	return &S3Client{
		client:        client,
		presignClient: presignClient,
		bucket:        bucket,
		tempPath:      tempPath,
	}, nil
}

func (s *S3Client) GenerateUploadURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	req, err := s.presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expiry))
	if err != nil {
		return "", fmt.Errorf("failed to generate upload URL: %w", err)
	}

	return req.URL, nil
}

func (s *S3Client) GenerateDownloadURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	req, err := s.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expiry))
	if err != nil {
		return "", fmt.Errorf("failed to generate download URL: %w", err)
	}

	return req.URL, nil
}

func (s *S3Client) DownloadToTempDir(ctx context.Context, key string) (string, error) {
	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return "", fmt.Errorf("failed to get object: %w", err)
	}
	defer result.Body.Close()

	now := time.Now().Format("20060102_150405")

	tmp, err := os.CreateTemp(s.tempPath, "video-input-*"+now+filepath.Ext(key))
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tmp.Close()

	if _, err = io.Copy(tmp, result.Body); err != nil {
		os.Remove(tmp.Name())
		return "", fmt.Errorf("failed to write temp file: %w", err)
	}

	return tmp.Name(), nil
}

func (s *S3Client) UploadFile(ctx context.Context, filePath, key string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open zip file: %w", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat zip file: %w", err)
	}

	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(key),
		Body:          file,
		ContentLength: aws.Int64(stat.Size()),
		ContentType:   aws.String("application/zip"),
	})
	if err != nil {
		return fmt.Errorf("failed to upload zip: %w", err)
	}

	return nil
}
