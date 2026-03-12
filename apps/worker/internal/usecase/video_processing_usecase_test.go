package usecase

import (
	"archive/zip"
	"context"
	"os"
	"path/filepath"
	"testing"
	"github.com/stretchr/testify/require"
)

func createTempFrame(t *testing.T, dir, name string) string {
	t.Helper()

	path := filepath.Join(dir, name)

	err := os.WriteFile(path, []byte("fake frame"), 0644)
	require.NoError(t, err)

	return path
}

func TestCreateZip_Success(t *testing.T) {
	tempDir := t.TempDir()

	frame1 := createTempFrame(t, tempDir, "frame_0001.png")
	frame2 := createTempFrame(t, tempDir, "frame_0002.png")

	outputZip := filepath.Join(tempDir, "frames.zip")

	usecase := NewVideoProcessingUseCase(tempDir, tempDir)

	err := usecase.createZip([]string{frame1, frame2}, outputZip)

	require.NoError(t, err)

	_, err = os.Stat(outputZip)
	require.NoError(t, err)

	zipReader, err := zip.OpenReader(outputZip)
	require.NoError(t, err)
	defer zipReader.Close()

	require.Len(t, zipReader.File, 2)
}

func TestAddFileToZip_Success(t *testing.T) {
	tempDir := t.TempDir()

	frame := createTempFrame(t, tempDir, "frame_0001.png")

	zipPath := filepath.Join(tempDir, "test.zip")

	zipFile, err := os.Create(zipPath)
	require.NoError(t, err)
	defer zipFile.Close() // <- IMPORTANTE

	zipWriter := zip.NewWriter(zipFile)

	usecase := NewVideoProcessingUseCase(tempDir, tempDir)

	err = usecase.addFileToZip(zipWriter, frame)
	require.NoError(t, err)

	err = zipWriter.Close()
	require.NoError(t, err)
}

func TestDeleteLocalFiles_Success(t *testing.T) {
	tempDir := t.TempDir()

	zipDir := filepath.Join(tempDir, "zip")
	framesDir := filepath.Join(tempDir, "frames")
	downloadDir := filepath.Join(tempDir, "download")

	os.MkdirAll(zipDir, 0755)
	os.MkdirAll(framesDir, 0755)
	os.MkdirAll(downloadDir, 0755)

	usecase := NewVideoProcessingUseCase(zipDir, framesDir)

	err := usecase.DeleteLocalFiles(zipDir, framesDir, downloadDir)

	require.NoError(t, err)

	_, err = os.Stat(zipDir)
	require.True(t, os.IsNotExist(err))

	_, err = os.Stat(framesDir)
	require.True(t, os.IsNotExist(err))
}

func TestExtractFrames_NoFrames(t *testing.T) {
	tempDir := t.TempDir()

	usecase := NewVideoProcessingUseCase(tempDir, tempDir)

	ctx := context.Background()

	_, _, err := usecase.extractFrames(ctx, "video.mp4")

	require.Error(t, err)
}

func TestProcess_NoFrames(t *testing.T) {
	tempDir := t.TempDir()

	usecase := NewVideoProcessingUseCase(tempDir, tempDir)

	ctx := context.Background()

	_, _, err := usecase.Process(ctx, "video.mp4")

	require.Error(t, err)
}

func TestCreateZip_FileNotFound(t *testing.T) {
	tempDir := t.TempDir()

	usecase := NewVideoProcessingUseCase(tempDir, tempDir)

	outputZip := filepath.Join(tempDir, "frames.zip")

	err := usecase.createZip([]string{"arquivo_inexistente.png"}, outputZip)

	require.Error(t, err)
}

func TestExtractFrames_Success(t *testing.T) {
	tempDir := t.TempDir()

	usecase := NewVideoProcessingUseCase(tempDir, tempDir)

	// caminho que extractFrames vai criar
	expectedDir := filepath.Join(tempDir, "video")

	err := os.MkdirAll(expectedDir, 0755)
	require.NoError(t, err)

	// cria frames fake que o Glob vai encontrar
	createTempFrame(t, expectedDir, "frame_0001.png")
	createTempFrame(t, expectedDir, "frame_0002.png")

	// cria ffmpeg fake executável (Linux)
	ffmpegPath := filepath.Join(tempDir, "ffmpeg")

	script := "#!/bin/sh\nexit 0\n"

	err = os.WriteFile(ffmpegPath, []byte(script), 0755)
	require.NoError(t, err)

	// adiciona ao PATH usando separador correto
	oldPath := os.Getenv("PATH")
	os.Setenv("PATH", tempDir+string(os.PathListSeparator)+oldPath)
	defer os.Setenv("PATH", oldPath)

	ctx := context.Background()

	frames, path, err := usecase.extractFrames(ctx, "video.mp4")

	require.NoError(t, err)
	require.Len(t, frames, 2)
	require.Equal(t, expectedDir, path)
}

func TestProcess_Success(t *testing.T) {
	tempDir := t.TempDir()

	usecase := NewVideoProcessingUseCase(tempDir, tempDir)

	expectedDir := filepath.Join(tempDir, "video")

	err := os.MkdirAll(expectedDir, 0755)
	require.NoError(t, err)

	createTempFrame(t, expectedDir, "frame_0001.png")
	createTempFrame(t, expectedDir, "frame_0002.png")

	// cria ffmpeg fake executável (Linux)
	ffmpegPath := filepath.Join(tempDir, "ffmpeg")

	script := "#!/bin/sh\nexit 0\n"

	err = os.WriteFile(ffmpegPath, []byte(script), 0755)
	require.NoError(t, err)

	// adiciona ao PATH corretamente
	oldPath := os.Getenv("PATH")
	os.Setenv("PATH", tempDir+string(os.PathListSeparator)+oldPath)
	defer os.Setenv("PATH", oldPath)

	ctx := context.Background()

	framesPath, zipPath, err := usecase.Process(ctx, "video.mp4")

	require.NoError(t, err)
	require.Equal(t, expectedDir, framesPath)

	// garante que o zip foi criado
	require.FileExists(t, zipPath)
}