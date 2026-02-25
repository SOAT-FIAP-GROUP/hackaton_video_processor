package usecase

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type VideoProcessingUseCase struct {
	zipTempDir    string
	framesTempDir string
}

func NewVideoProcessingUseCase(framesDir, zipDir string) *VideoProcessingUseCase {
	return &VideoProcessingUseCase{
		zipTempDir:    zipDir,
		framesTempDir: framesDir,
	}
}

func (v *VideoProcessingUseCase) Process(videoKey string) (string, string, error) {
	frames, framesPath, err := v.extractFrames(videoKey)
	if err != nil {
		return "", "", fmt.Errorf("error at video processing: %v", err)
	}

	if len(frames) == 0 {
		return "", "", fmt.Errorf("no frames found for key: %s", videoKey)
	}

	timestamp := time.Now().Format("20060102_150405")
	zipFilename := fmt.Sprintf("frames_%s.zip", timestamp)
	zipPath := filepath.Join(v.zipTempDir, zipFilename)

	err = v.createZip(frames, zipPath)

	return framesPath, zipPath, nil
}

func (v *VideoProcessingUseCase) extractFrames(videoPath string) ([]string, string, error) {
	keySplit := strings.Split(videoPath, "/")
	fileName := keySplit[len(keySplit)-1]
	fileNameWithoutExtension := strings.Split(fileName, ".")[0]

	zipPath := filepath.Join(v.zipTempDir, fileNameWithoutExtension)
	framePattern := filepath.Join(zipPath, "frame_%04d.png")

	err := os.MkdirAll(zipPath, 0755)
	if err != nil {
		return nil, "", fmt.Errorf("error at extracting frames: %v", err)
	}

	cmd := exec.Command("ffmpeg",
		"-i", videoPath,
		"-vf", "fps=1",
		"-y",
		framePattern,
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, "", fmt.Errorf("error extraction frames: %s", out)
	}

	frames, err := filepath.Glob(filepath.Join(zipPath, "*.png"))
	if err != nil || len(frames) == 0 {
		return nil, "", fmt.Errorf("nenhum frame foi extraído do vídeo")
	}

	return frames, zipPath, nil
}

func (v *VideoProcessingUseCase) createZip(frames []string, outputPath string) error {
	zipFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("erro ao criar arquivo ZIP: %w", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	for _, frame := range frames {
		err := v.addFileToZip(zipWriter, frame)
		if err != nil {
			return fmt.Errorf("erro ao adicionar arquivo ao ZIP: %w", err)
		}
	}

	return nil
}

func (v *VideoProcessingUseCase) addFileToZip(zipWriter *zip.Writer, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	header.Name = filepath.Base(filename)
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, file)
	return err
}

func (v *VideoProcessingUseCase) DeleteLocalFiles(zip, frames, download string) error {
	err := os.RemoveAll(zip)
	if err != nil {
		return fmt.Errorf("error at deleting zip file: %v", err)
	}
	err = os.RemoveAll(download)
	if err != nil {
		return fmt.Errorf("error at deleting download file: %v", err)
	}

	err = os.RemoveAll(frames)
	if err != nil {
		return fmt.Errorf("error at deleting frames: %v", err)
	}

	return nil
}
