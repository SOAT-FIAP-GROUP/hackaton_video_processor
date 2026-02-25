package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"shared/storage/S3"
)

func UploadViaSignedURL(signedURL, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Get file size so Go doesn't use chunked transfer encoding
	stat, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	req, err := http.NewRequest(http.MethodPut, signedURL, file)
	if err != nil {
		return err
	}

	req.ContentLength = stat.Size() // <- this is the key fix

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("upload failed with status: %d, body: %s", resp.StatusCode, body)
	}

	return nil
}

// DownloadViaSignedURL downloads a file using a pre-signed URL (client-side)
func DownloadViaSignedURL(signedURL, destPath string) error {
	resp, err := http.Get(signedURL)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	file, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	return err
}

func main() {
	ctx := context.Background()

	s3Client, err := S3.NewS3Client(ctx, "upload-test-roger", "us-east-1")
	if err != nil {
		log.Fatal(err)
	}

	// Generate upload URL (valid for 15 minutes)
	uploadURL, err := s3Client.GenerateUploadURL(ctx, "uploads/Gravação de tela de 14-10-2025 18:33:33.webm", 1*time.Minute)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Upload URL:", uploadURL)

	err = UploadViaSignedURL(uploadURL, "/home/roger/Vídeos/Gravações de tela/Gravação de tela de 14-10-2025 18:33:33.webm")
	if err != nil {
		panic("failed to upload file: " + err.Error())
	}

	// Generate download URL (valid for 1 hour)
	downloadURL, err := s3Client.GenerateDownloadURL(ctx, "uploads/Gravação de tela de 14-10-2025 18:33:33.webm", 1*time.Hour)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Download URL:", downloadURL)
}
