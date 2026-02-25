package domain

import "mime/multipart"

// VideoRepository defines the interface for video storage operations
type VideoRepository interface {
	Save(file multipart.File, filename string) (string, int64, error)
	Delete(path string) error
	IsValidFormat(filename string) bool
}

// FrameRepository defines the interface for frame storage operations
type FrameRepository interface {
	CreateZip(frames []string, outputPath string) error
	ListZipFiles() ([]FileInfo, error)
	GetZipPath(filename string) (string, error)
}

// VideoProcessor defines the interface for video processing operations
type VideoProcessor interface {
	ExtractFrames(videoPath, outputDir string) ([]string, error)
	CleanupTempDir()
}

type UserRepository interface {
	Signup(user *User) error
	Login(email, password string) (*User, error)
	FindByEmail(email string) (*User, error)
}

type VideoEntityRepository interface {
	Save(video *Video) (*Video, error)
	FindByUserId(userId string) ([]*Video, error)
	FindById(videoId string) (*Video, error)
	Update(video *Video) (*Video, error)
	ListByStatus(processed bool, userId string) ([]*Video, error)
}
