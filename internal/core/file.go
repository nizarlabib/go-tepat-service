package core

import (
	"context"
	"fmt"
	"net/http"
	"slices"

	"go-memolab-service/internal/entity"
	"go-memolab-service/internal/framework/secondary/repository"
	"go-memolab-service/internal/framework/secondary/storage"
	"go-memolab-service/internal/spesial"
	"go-memolab-service/internal/value_object/v1/constant"
	"go-memolab-service/internal/value_object/v1/convert"
	"go-memolab-service/internal/value_object/v1/handling"
)

type File interface {
	SumSizeByUploaderID(ctx context.Context, uID uint) (int64, error)

	FindByID(ctx context.Context, e *entity.File) error
	FindByName(ctx context.Context, e *entity.File) error
	FindByNames(ctx context.Context, names []string) ([]*entity.File, error)

	Create(ctx context.Context, e *entity.File) error
	Update(ctx context.Context, e *entity.File) error
	Delete(ctx context.Context, e *entity.File) error

	GetFileObject(ctx context.Context, e *entity.File) ([]byte, error)
	UploadFiles(ctx context.Context, fs []*entity.FileUpload) error
}

type fileOption func(fac *file)

func NewFile(opts ...fileOption) File {
	fac := &file{}

	for _, opt := range opts {
		opt(fac)
	}

	return fac
}

func FileWithFileRepo(repo repository.File) fileOption {
	return func(fac *file) {
		fac.repo = repo
	}
}

func FileWithStorage(storage Storage) fileOption {
	return func(fac *file) {
		fac.storage = storage
	}
}

type file struct {
	repo    repository.File
	storage Storage
}

func (c *file) SumSizeByUploaderID(ctx context.Context, uID uint) (int64, error) {
	return c.repo.SumSizeByUploaderID(ctx, uID)
}

func (c *file) FindByID(ctx context.Context, e *entity.File) error {
	return c.repo.FindByID(ctx, e)
}

func (c *file) FindByName(ctx context.Context, e *entity.File) error {
	return c.repo.FindByName(ctx, e)
}

func (c *file) FindByNames(ctx context.Context, names []string) ([]*entity.File, error) {
	return c.repo.FindByNames(ctx, names)
}

func (c *file) Create(ctx context.Context, e *entity.File) error {
	return c.repo.Create(ctx, e)
}

func (c *file) Update(ctx context.Context, e *entity.File) error {
	return c.repo.Update(ctx, e)
}

func (c *file) Delete(ctx context.Context, e *entity.File) error {
	var err error
	if err = c.repo.Delete(ctx, e); err != nil {
		return err
	}

	_ = c.storage.DeleteFileByKey(context.Background(), fmt.Sprintf("%s/%s", e.Path, e.Name))
	return nil
}

func (c *file) UploadFiles(ctx context.Context, fs []*entity.FileUpload) error {
	user, _ := spesial.GetUserFromContext(ctx)

	var fileMap *entity.FileMap
	var accFileSize int64
	var accFileSizeLimit int64
	var files []*entity.File
	var fileObjects []*storage.FileObject

	accFileSize = 0
	accFileSizeLimit = 0

	for _, f := range fs {

		if fileMap == nil {
			fileMap = getFileValidationFromPath(f.Path)
		}

		if accFileSizeLimit == 0 {
			accFileSizeLimit = int64(fileMap.LimitInMB * (1 << 20))
		}

		fileSize := f.MFileHeader.Size
		contentType := f.MFileHeader.Header.Get("Content-Type")

		fileExt := getFileExtFromMime(contentType)

		if !entity.FileTypeContains(fileMap.FileType, fileExt) {
			return handling.NewHttpError(nil, http.StatusBadRequest, fmt.Sprintf("file: %s format is not supported", f.Field), constant.ERR_INVALID_FILE_FORMAT)
		}

		accFileSize += fileSize

		if !fileMap.IsAccumulated && fileSize > accFileSizeLimit {
			return handling.NewHttpError(nil, http.StatusBadRequest, fmt.Sprintf("file: %s size is too big", f.Field), constant.ERR_FILE_SIZE_TOO_BIG)
		}

		file := &entity.File{
			Name: convert.GenerateFileNameWithExt(fileExt),
			Path: f.Path,
			Mime: contentType,
			Size: fileSize,
			Tags: []string{
				"memolab-service",
				f.Path,
			},
			UploaderID: user.ID,
		}
		files = append(files, file)

		fileObject := &storage.FileObject{
			Name: fmt.Sprintf("%s/%s", file.Path, file.Name),
			Size: fileSize,
			File: storage.FileFromMultipartHeader(f.MFileHeader),
		}
		fileObjects = append(fileObjects, fileObject)

		f.ObjectName = file.Name
		f.Size = fileSize
	}

	if fileMap.IsAccumulated && accFileSize != 0 && accFileSize > accFileSizeLimit {
		return handling.NewHttpError(nil, http.StatusBadRequest, "file size is too big", constant.ERR_FILE_SIZE_TOO_BIG)
	}

	todayUsage, err := c.SumSizeByUploaderID(ctx, user.ID)
	if err != nil {
		return err
	}

	if (todayUsage + accFileSize) > user.DiskQuota {
		return handling.NewHttpError(nil, http.StatusUnprocessableEntity, "user disk's quota is exceeded", constant.ERR_USER_QUOTA_EXCEEDED)
	}

	uploadCtx, cancel := context.WithCancel(context.Background())
	if err := c.storage.UploadFiles(uploadCtx, fileObjects); err != nil {
		cancel()
		return err
	}

	if err := c.repo.Create(ctx, files...); err != nil {
		cancel()
		c.storage.DeleteFiles(ctx, fileObjects)
		return err
	}

	return nil
}

func (c *file) GetFileObject(ctx context.Context, e *entity.File) ([]byte, error) {
	if e.Path == "" {
		if e.ID != "" {
			if err := c.FindByID(ctx, e); err != nil {
				return nil, err
			}
		} else if e.Name != "" {
			if err := c.FindByName(ctx, e); err != nil {
				return nil, err
			}
		}
	}

	objectName := fmt.Sprintf("%s/%s", e.Path, e.Name)
	return c.storage.GetFile(ctx, objectName)
}

func getFileValidationFromPath(path string) *entity.FileMap {
	idx := slices.IndexFunc(entity.ValidFilesMapping, func(m entity.FileMap) bool {
		return m.Path == path
	})
	if idx < 0 {
		return nil
	}
	return &entity.ValidFilesMapping[idx]
}

func getFileExtFromMime(mime string) string {
	switch mime {
	case constant.FileTypePDF:
		return "pdf"
	case constant.FileTypeImageJpeg:
		return "jpeg"

	case constant.FileTypeImagePng:
		return "png"
	case constant.FileTypeImageJpg:
		return "jpg"
	}
	return ""
}

func (c *file) GetFileByName(ctx context.Context, filename string) ([]byte, error) {
	e := &entity.File{Name: filename}
	buff, err := c.GetFileObject(ctx, e)
	if err != nil {
		return nil, err
	}

	return buff, nil
}
