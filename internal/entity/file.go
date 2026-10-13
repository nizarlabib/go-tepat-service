package entity

import (
	"fmt"
	"mime/multipart"
	"net/http"
)

type File struct {
	BaseDeletableModel
	UploaderID uint     `json:"uploader_id" gorm:"not_null"`
	Name       string   `json:"name" gorm:"not null;type:varchar(100);"`
	Path       string   `json:"path" gorm:"not null;type:varchar(50);"`
	Size       int64    `json:"size" gorm:"not null;default:0;"`
	Mime       string   `json:"mime" gorm:"not null;type:varchar(100);"`
	Tags       []string `json:"tags" gorm:"type:jsonb;serializer:json;"`
}

type FileUpload struct {
	FileInfo    *File
	ObjectName  string
	Size        int64
	Path        string
	Field       string
	MFileHeader *multipart.FileHeader
	MFile       multipart.File
}

type FileMap struct {
	FileType      []string
	Path          string
	LimitInMB     float64
	IsAccumulated bool
}

var ValidFilesMapping = []FileMap{
	{FileType: []string{"pdf", "png", "jpeg", "jpg"}, Path: "profile_picture", LimitInMB: 25, IsAccumulated: false},
}

func FileTypeContains(arr []string, str string) bool {
	for _, item := range arr {
		if item == str {
			return true
		}
	}
	return false
}

func ParseFileUpload(r *http.Request, field string) (*FileUpload, error) {
	// Parse multipart form (max 10MB)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		return nil, fmt.Errorf("failed to parse multipart form: %w", err)
	}

	fileHeader, ok := r.MultipartForm.File[field]
	if !ok || len(fileHeader) == 0 {
		return nil, fmt.Errorf("file not found for field: %s", field)
	}

	fh := fileHeader[0]

	return &FileUpload{
		ObjectName:  fh.Filename,
		Size:        fh.Size,
		Path:        "",
		Field:       field,
		MFileHeader: fh,
	}, nil
}
