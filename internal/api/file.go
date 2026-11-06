package api

import (
	"context"

	"go-starter-service/internal/core"
	"go-starter-service/internal/entity"
)

type File interface {
	UploadFiles(ctx context.Context, e []*entity.FileUpload) error
	GetFileObject(ctx context.Context, e *entity.File) ([]byte, error)
	Delete(ctx context.Context, e *entity.File) error
}

func NewFile(core core.File) File {
	return &fileApi{
		core: core,
	}
}

type fileApi struct {
	core core.File
}

func (a *fileApi) UploadFiles(ctx context.Context, e []*entity.FileUpload) error {
	return a.core.UploadFiles(ctx, e)
}

func (a *fileApi) GetFileObject(ctx context.Context, e *entity.File) ([]byte, error) {
	return a.core.GetFileObject(ctx, e)
}

func (a *fileApi) Delete(ctx context.Context, e *entity.File) error {
	return a.core.Delete(ctx, e)
}
