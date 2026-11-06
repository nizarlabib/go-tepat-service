package storage

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net/http"

	"go-starter-service/internal/infrastructure"
	"go-starter-service/internal/pkg/encryption"
	"go-starter-service/internal/value_object/v1/constant"
	"go-starter-service/internal/value_object/v1/handling"

	"github.com/minio/minio-go/v7"
	"github.com/spf13/viper"
)

var ErrReadFile = errors.New("read multipart.file error")

func NewMinio(infra infrastructure.Minio) Storage {
	return &minioStorage{
		client: infra.GetClient(),
		bucket: infra.GetBucketName(),
		aesKey: viper.GetString("AES_KEY"),
	}
}

type minioStorage struct {
	client *minio.Client
	bucket string
	aesKey string
}

func (s *minioStorage) DeleteFileByKey(ctx context.Context, key string) error {
	err := s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{ForceDelete: true})
	if err != nil {
		fmt.Printf("error occured: %v\n", err)
		return err
	}
	return nil
}

func (s *minioStorage) uploadFile(ctx context.Context, e *FileObject) error {
	if e.Name == "" {
		return handling.NewHttpError500(nil)
	}
	reader, err := e.File.Open()
	if err != nil {
		return handling.NewHttpError500(err)
	}
	defer reader.Close()

	aesBlock, err := encryption.CreateAesBlock(s.aesKey)
	if err != nil {
		return handling.NewHttpError500(err)
	}
	encryptedText, err := encryption.EncryptFile(reader, aesBlock)

	e.Size = int64(binary.Size(encryptedText))

	if err != nil {
		return err
	}
	err = s.uploadEncryptedText(ctx, encryptedText, e.Name, e.Size, e.ContentType)
	if err != nil {
		return err
	}
	return nil
}

func (s *minioStorage) UploadFiles(ctx context.Context, e []*FileObject) error {
	for _, file := range e {
		if err := s.uploadFile(ctx, file); err != nil {
			return err
		}
	}
	return nil
}

func (s *minioStorage) DeleteFile(ctx context.Context, e *FileObject) error {
	if e.Name == "" {
		return handling.NewHttpError500(nil)
	}
	if err := s.DeleteFileByKey(ctx, e.Name); err != nil {
		return err
	}
	return nil
}

func (s *minioStorage) DeleteFiles(ctx context.Context, e []*FileObject) error {
	for _, file := range e {
		if err := s.DeleteFile(ctx, file); err != nil {
			return err
		}
	}
	return nil
}

func (s *minioStorage) FileExists(ctx context.Context, key string) (bool, error) {
	_, err := s.client.StatObject(ctx, s.bucket, key, minio.StatObjectOptions{})
	if err != nil {
		errResp := minio.ToErrorResponse(err)
		if errResp.Code == "NoSuchKey" || errResp.StatusCode == http.StatusNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *minioStorage) GetFile(ctx context.Context, key string) ([]byte, error) {
	_, err := s.client.StatObject(ctx, s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, handling.NewHttpError404(err, "file is not found")
	}

	file, err := s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, handling.NewHttpError(err, http.StatusUnprocessableEntity, "error retrieving file", constant.ERR_UNKNOWN)
	}

	bytes, err := io.ReadAll(file)
	if err != nil {
		// GetObject bersifat lazy: object yang tidak ada baru terdeteksi saat dibaca.
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return nil, handling.NewHttpError404(err, "file is not found")
		}
		return nil, handling.NewHttpError500(err)
	}

	aesBlock, err := encryption.CreateAesBlock(s.aesKey)
	if err != nil {
		return nil, handling.NewHttpError500(err)
	}
	decrypted, err := encryption.Decrypt(bytes, aesBlock)
	if err != nil {
		return nil, handling.NewHttpError(err, http.StatusUnprocessableEntity, "error decrypting file", constant.ERR_UNKNOWN)
	}

	return decrypted, nil
}

func (s *minioStorage) uploadEncryptedText(ctx context.Context, encryptedText []byte, objectName string, objectSize int64, objectContentType string) error {
	if objectContentType == "" {
		objectContentType = "application/octet-stream"
	}
	_, err := s.client.PutObject(ctx, s.bucket, objectName, bytes.NewReader(encryptedText), objectSize, minio.PutObjectOptions{
		ContentType: "application/octet-stream",
		UserTags: map[string]string{
			"service": "go-starter-service",
		},
	})
	if err != nil {
		return fmt.Errorf("failed to upload to MinIO: %w", err)
	}
	return nil
}
