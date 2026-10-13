package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go-memolab-service/internal/entity"
	"go-memolab-service/internal/value_object/v1/handling"

	"gorm.io/gorm"
)

type File interface {
	GetDefaultProfilePicture(ctx context.Context) (*entity.File, error)
	SumSizeByUploaderID(ctx context.Context, uID uint) (int64, error)
	FindByID(ctx context.Context, e *entity.File) error
	FindByName(ctx context.Context, e *entity.File) error
	FindByNames(ctx context.Context, names []string) ([]*entity.File, error)
	Create(ctx context.Context, e ...*entity.File) error
	Update(ctx context.Context, e *entity.File) error
	Delete(ctx context.Context, e *entity.File) error
}

func NewFile(db *gorm.DB) File {
	return &file{
		db: db,
	}
}

type file struct {
	db *gorm.DB
}

func (r *file) GetDefaultProfilePicture(ctx context.Context) (*entity.File, error) {
	var file *entity.File
	err := r.db.WithContext(ctx).Where("name = ?", "default_profile_picture.png").First(&file).Error
	if err != nil {
		return nil, err
	}
	return file, nil
}

func (r *file) SumSizeByUploaderID(ctx context.Context, uID uint) (int64, error) {
	var totalSize int64
	if uID == 0 {
		return 0, nil
	}

	err := r.db.WithContext(ctx).Raw("SELECT COALESCE(SUM(size), 0) AS total_usage FROM files WHERE uploader_id = ? AND created_at::date = CURRENT_DATE", uID).Scan(&totalSize).Error
	if err != nil {
		return 0, nil
	}

	return 0, nil
}

func (r *file) FindByID(ctx context.Context, e *entity.File) error {
	if e.ID == "" {
		return handling.NewHttpError404(nil, "file not found")
	}
	if err := r.db.WithContext(ctx).First(e).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return handling.NewHttpError404(err, "file not found")
		}
		return handling.NewHttpError500(err)
	}
	return nil
}

func (r *file) FindByName(ctx context.Context, e *entity.File) error {
	if e.Name == "" {
		return handling.NewHttpError404(nil, "file not found")
	}
	if err := r.db.WithContext(ctx).Where("name = ?", e.Name).First(e).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return handling.NewHttpError404(err, fmt.Sprintf("file with name %s is not found", e.Name))
		}
		return handling.NewHttpError500(err)
	}
	return nil
}

func (r *file) FindByNames(ctx context.Context, names []string) ([]*entity.File, error) {
	if len(names) == 0 {
		return nil, handling.NewHttpError404(nil, "file not found")
	}
	var files []*entity.File
	if err := r.db.WithContext(ctx).Where("name in ?", names).Find(&files).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, handling.NewHttpError404(err, "file not found")
		}
		return nil, handling.NewHttpError500(err)
	}
	return files, nil
}

func (r *file) Create(ctx context.Context, e ...*entity.File) error {
	if err := r.db.WithContext(ctx).Create(e).Error; err != nil {
		return handling.NewHttpError500(err)
	}
	return nil
}

func (r *file) Delete(ctx context.Context, e *entity.File) error {
	if e.Name == "" {
		return handling.NewHttpError404(nil, "file name is required")
	}
	if err := r.db.WithContext(ctx).Delete(&entity.File{}, "name = ?", e.Name).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return handling.NewHttpError404(err, "file not found")
		}
		return handling.NewHttpError500(err)
	}
	return nil
}

func (r *file) Update(ctx context.Context, e *entity.File) error {
	tx := r.db.WithContext(ctx).Begin()
	e.UpdatedAt = time.Now()

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := tx.Model(entity.File{}).Where("id = ?", e.ID).Updates(e).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			return handling.NewHttpError404(err, fmt.Sprintf("file with id %s is not found", e.ID))
		}
		return handling.NewHttpError500(err)
	}

	tx.Commit()
	return nil
}
