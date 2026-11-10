package repository

import (
	"context"
	"go-starter-service/internal/entity"
	"go-starter-service/internal/infrastructure"
	"go-starter-service/internal/pkg/dd"
	"go-starter-service/internal/value_object/v1/handling"
	"go-starter-service/internal/value_object/v1/pagination"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Category interface {
	Create(ctx context.Context, e *entity.Category) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Category, error)
	FindBySlug(ctx context.Context, slug string) (*entity.Category, error)
	FindAll(ctx context.Context, e *entity.Category, pg *pagination.Pagination) error
	Update(ctx context.Context, e *entity.Category) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindCategories(ctx context.Context) ([]*entity.Category, error)
	CategoryExists(ctx context.Context, id uuid.UUID) (bool, error)
}

func NewCategory(infra infrastructure.Postgres) Category {
	db := infra.GetClientGorm()
	return &category{db: db}
}

type category struct {
	db *gorm.DB
}

func (r *category) Create(ctx context.Context, e *entity.Category) error {
	tx := r.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()
	defer tx.Rollback()

	dd.PrettyPrint(e)
	result := tx.Create(e)
	if result.Error != nil {
		return handling.NewHttpError500(result.Error)
	}
	if result.RowsAffected == 0 {
		return handling.NewHttpError500(nil)
	}

	tx.Commit()
	return nil
}

func (r *category) FindByID(ctx context.Context, id uuid.UUID) (*entity.Category, error) {
	var category *entity.Category
	result := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&category)
	if result.Error != nil {
		return nil, handling.NewHttpError500(result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, handling.NewHttpError404(nil, "category not found")
	}
	return category, nil
}

func (r *category) FindBySlug(ctx context.Context, slug string) (*entity.Category, error) {
	var category *entity.Category
	result := r.db.WithContext(ctx).Where("slug = ? AND deleted_at IS NULL", slug).First(&category)
	if result.Error != nil {
		return nil, handling.NewHttpError500(result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, handling.NewHttpError404(nil, "category not found")
	}
	return category, nil
}

func (r *category) FindAll(ctx context.Context, e *entity.Category, pg *pagination.Pagination) error {
	var categories []*entity.Category

	var conditions []string
	var values []interface{}

	conditions = append(conditions, "deleted_at IS NULL")

	if e.Name != "" {
		conditions = append(conditions, "name ILIKE ?")
		values = append(values, "%"+e.Name+"%")
	}

	query := strings.Join(conditions, " AND ")
	args := append([]interface{}{query}, values...)

	db := r.db.WithContext(ctx).
		Scopes(pagination.Paginate(&entity.Category{}, pg, r.db, args...)).
		Find(&categories)

	if db.Error != nil {
		return handling.NewHttpError500(db.Error)
	}

	pg.Rows = categories
	return nil
}

func (r *category) Update(ctx context.Context, e *entity.Category) error {
	tx := r.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()
	defer tx.Rollback()

	result := tx.Save(e)
	if result.Error != nil {
		return handling.NewHttpError500(result.Error)
	}
	if result.RowsAffected == 0 {
		return handling.NewHttpError404(nil, "category not found")
	}

	tx.Commit()
	return nil
}

func (r *category) FindCategories(ctx context.Context) ([]*entity.Category, error) {
	var categories []*entity.Category
	result := r.db.WithContext(ctx).
		Where("is_active = ? AND deleted_at IS NULL", true).
		Order("sort_order ASC, name ASC").
		Find(&categories)
	if result.Error != nil {
		return nil, handling.NewHttpError500(result.Error)
	}
	return categories, nil
}

func (r *category) CategoryExists(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).
		Model(&entity.Category{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Count(&count)
	if result.Error != nil {
		return false, handling.NewHttpError500(result.Error)
	}
	return count > 0, nil
}

func (r *category) Delete(ctx context.Context, id uuid.UUID) error {
	tx := r.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()
	defer tx.Rollback()

	result := tx.Where("id = ? AND deleted_at IS NULL", id).Delete(&entity.Category{})
	if result.Error != nil {
		return handling.NewHttpError500(result.Error)
	}
	if result.RowsAffected == 0 {
		return handling.NewHttpError404(nil, "category not found")
	}

	tx.Commit()
	return nil
}
