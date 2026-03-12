package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"shared/SQS"
	"shared/database/repository"
	"shared/storage/S3"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type VideoHandler struct {
	fileStore *S3.S3Client
	emitter   *SQS.SQSEmitter
	videoRepo *repository.VideoRepository
}

func NewVideoHandler(s3client *S3.S3Client, emitter *SQS.SQSEmitter, repo *repository.VideoRepository) *VideoHandler {
	return &VideoHandler{
		fileStore: s3client,
		emitter:   emitter,
		videoRepo: repo,
	}
}

type UploadResponse struct {
	Success    bool     `json:"success"`
	Message    string   `json:"message"`
	ZipPath    string   `json:"zip_path,omitempty"`
	FrameCount int      `json:"frame_count,omitempty"`
	Images     []string `json:"images,omitempty"`
}

func (h *VideoHandler) HandleUpload(c *gin.Context) {
	file, header, err := c.Request.FormFile("video")
	if err != nil {
		c.JSON(http.StatusBadRequest, UploadResponse{
			Success: false,
			Message: "Erro ao receber arquivo: " + err.Error(),
		})
		return
	}
	defer file.Close()

	userID := c.GetString("userID")
	username := c.GetString("name")
	userEmail := c.GetString("email")

	log.Printf("Usuário logado: %s", userID)

	filename := header.Filename
	filename = fmt.Sprintf("%s/%s/%s", "uploads", userID, filename)

	log.Printf("Filename: %s", filename)

	ctx := c.Request.Context()
	fileSize := header.Size
	fileContent := make([]byte, fileSize)
	_, err = file.Read(fileContent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, UploadResponse{
			Success: false,
			Message: "Erro ao ler arquivo: " + err.Error(),
		})
		return
	}

	uploadUrl, err := h.fileStore.GenerateUploadURL(ctx, filename, 10*time.Second)
	if err != nil {
		c.JSON(http.StatusInternalServerError, UploadResponse{
			Success: false,
			Message: "Erro ao gerar URL de upload: " + err.Error(),
		})
		return
	}

	err = uploadViaSignedURL(uploadUrl, fileContent, fileSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, UploadResponse{
			Success: false,
			Message: "Erro ao fazer upload do arquivo: " + err.Error(),
		})
		return
	}

	brokerMessage := SQS.BrokerMessage{
		VideoPath: filename,
		UserID:    userID,
		UserName:  username,
		UserEmail: userEmail,
		UploadAt:  time.Now(),
	}

	jsonMessage, err := json.Marshal(brokerMessage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, UploadResponse{
			Success: false,
			Message: "Erro ao ler arquivo: " + err.Error(),
		})
	}

	messageID, err := h.emitter.SendMessage(ctx, jsonMessage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, UploadResponse{
			Success: false,
			Message: "Erro ao ler arquivo: " + err.Error(),
		})
	}

	log.Println("Message Sent:", messageID)

	c.JSON(http.StatusOK, UploadResponse{
		Success: true,
		Message: "Arquivo recebido com sucesso. Processamento em andamento.",
	})
}

func uploadViaSignedURL(signedURL string, file []byte, fileSize int64) error {
	reader := bytes.NewReader(file)

	req, err := http.NewRequest(http.MethodPut, signedURL, reader)
	if err != nil {
		return err
	}

	req.ContentLength = fileSize

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("Erro ao fazer upload via URL assinada:", err)
		return fmt.Errorf("upload failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("upload failed with status: %d, body: %s", resp.StatusCode, body)
	}

	return nil
}

func (h *VideoHandler) HandleDownload(c *gin.Context) {
	filename := c.Param("filename")
	cleanFileName := strings.TrimPrefix(filename, "/")
	userName := c.GetString("userID")
	filepath := fmt.Sprintf("downloads/%s/%s", userName, cleanFileName)

	url, err := h.fileStore.GenerateDownloadURL(c.Request.Context(), filepath, 5*time.Minute)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao ler arquivo: " + err.Error()})
	}

	c.Header("Content-Type", "application/json")

	c.JSON(http.StatusOK, gin.H{"download_url": url})
}

func (h *VideoHandler) HandleStatus(c *gin.Context) {
	userId := c.GetString("userID")

	files, err := h.videoRepo.ListVideosByUserID(c.Request.Context(), userId)
	if err != nil {
		log.Println("Erro ao listar arquivos processados:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao listar arquivos"})
		return
	}

	var results []map[string]interface{}
	for _, file := range files {
		pathSplit := strings.Split(file.Path, "/")
		downloadPath := pathSplit[len(pathSplit)-1]
		results = append(results, map[string]interface{}{
			"filename":     file.Name,
			"size":         0,
			"created_at":   file.ProcessedAt.Format("2006-01-02 15:04:05"),
			"download_url": downloadPath,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"files": results,
		"total": len(results),
	})
}

func (h *VideoHandler) HandleHome(c *gin.Context) {
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, GetHTMLForm())
}
