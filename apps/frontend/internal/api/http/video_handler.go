package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"shared/SQS"
	"shared/storage/S3"
	"time"

	"github.com/gin-gonic/gin"
)

type VideoHandler struct {
	fileStore *S3.S3Client
	emitter   *SQS.SQSEmitter
}

func NewVideoHandler(s3client *S3.S3Client, emitter *SQS.SQSEmitter) *VideoHandler {
	return &VideoHandler{
		fileStore: s3client,
		emitter:   emitter,
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

	err := fmt.Errorf("file not found: %s", filename)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Arquivo não encontrado"})
		return
	}

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "application/zip")

	c.File("filePath")
}

func (h *VideoHandler) HandleStatus(c *gin.Context) {
	//userId := c.GetString("userID")

	//files, err := h.videoUseCase.GetProcessedFiles(userId)
	err := fmt.Errorf("Simulated error for testing")
	if err != nil {
		log.Println("Erro ao listar arquivos processados:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao listar arquivos"})
		return
	}

	/*var results []map[string]interface{}
	for _, file := range files {
		results = append(results, map[string]interface{}{
			"filename":     file.Filename,
			"size":         file.Size,
			"created_at":   file.CreatedAt.Format("2006-01-02 15:04:05"),
			"download_url": file.DownloadURL,
		})
	}*/

	/*c.JSON(http.StatusOK, gin.H{
		"files": results,
		"total": len(results),
	})*/
}

func (h *VideoHandler) HandleHome(c *gin.Context) {
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, GetHTMLForm())
}
