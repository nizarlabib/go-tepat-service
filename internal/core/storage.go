package core

import (
	"context"

	"go-starter-service/internal/framework/secondary/storage"
)

type Storage interface {
	UploadFiles(ctx context.Context, e []*storage.FileObject) error
	DeleteFiles(ctx context.Context, e []*storage.FileObject) error
	DeleteFileByKey(ctx context.Context, key string) error
	GetFile(ctx context.Context, key string) ([]byte, error)
	FileExists(ctx context.Context, key string) (bool, error)
}

type StorageOption func(fac *storageCore)

func NewStorage(opts ...StorageOption) Storage {
	fac := &storageCore{}
	for _, opt := range opts {
		opt(fac)
	}
	return fac
}

func StorageWithStorage(storage storage.Storage) StorageOption {
	return func(fac *storageCore) {
		fac.storage = storage
	}
}

type storageCore struct {
	storage storage.Storage
}

// DeleteFileByKey implements Storage.
func (s *storageCore) DeleteFileByKey(ctx context.Context, key string) error {
	return s.storage.DeleteFileByKey(ctx, key)
}

func (s *storageCore) UploadFiles(ctx context.Context, e []*storage.FileObject) error {
	return s.storage.UploadFiles(ctx, e)
}

func (s *storageCore) DeleteFiles(ctx context.Context, e []*storage.FileObject) error {
	return s.storage.DeleteFiles(ctx, e)
}

func (s *storageCore) GetFile(ctx context.Context, key string) ([]byte, error) {
	return s.storage.GetFile(ctx, key)
}

func (s *storageCore) FileExists(ctx context.Context, key string) (bool, error) {
	return s.storage.FileExists(ctx, key)
}
