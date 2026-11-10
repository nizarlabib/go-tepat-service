package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"go-starter-service/internal/entity"
	"go-starter-service/internal/infrastructure"
	"go-starter-service/internal/value_object/v1/handling"
	"go-starter-service/internal/value_object/v1/pagination"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Product interface {
	Create(ctx context.Context, e *entity.Product) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Product, error)
	FindBySlug(ctx context.Context, slug string) (*entity.Product, error)
	FindAll(ctx context.Context, e *entity.Product, pg *pagination.Pagination) error
	Update(ctx context.Context, e *entity.Product) error
	UpdateStatus(ctx context.Context, id uuid.UUID, isActive bool) error
	Delete(ctx context.Context, id uuid.UUID) error
	SlugExists(ctx context.Context, slug string, excludeID uuid.UUID) (bool, error)
}

func NewProduct(infra infrastructure.Postgres) Product {
	db := infra.GetClientGorm()
	return &product{db: db}
}

type product struct {
	db *gorm.DB
}

func (r *product) scopedActiveNotDeleted(db *gorm.DB) *gorm.DB {
	return db.Where("is_active = ? AND deleted_at IS NULL", true)
}

func (r *product) Create(ctx context.Context, e *entity.Product) error {
	tx := r.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()
	defer tx.Rollback()

	result := tx.Create(e)
	if result.Error != nil {
		return handling.NewHttpError500(result.Error)
	}
	if result.RowsAffected == 0 {
		return handling.NewHttpError500(fmt.Errorf("failed to create product"))
	}

	tx.Commit()
	return nil
}

func (r *product) FindByID(ctx context.Context, id uuid.UUID) (*entity.Product, error) {
	var e entity.Product
	result := r.db.WithContext(ctx).
		Preload("Category").
		Where("id = ? AND deleted_at IS NULL", id).
		First(&e)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, handling.NewHttpError404(nil, "Produk tidak ditemukan.")
		}
		return nil, handling.NewHttpError500(result.Error)
	}
	return &e, nil
}

func (r *product) FindBySlug(ctx context.Context, slug string) (*entity.Product, error) {
	var e entity.Product
	result := r.db.WithContext(ctx).
		Preload("Category").
		Scopes(r.scopedActiveNotDeleted).
		Where("slug = ?", slug).
		First(&e)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, handling.NewHttpError404(nil, "Produk tidak ditemukan.")
		}
		return nil, handling.NewHttpError500(result.Error)
	}
	return &e, nil
}

func (r *product) FindAll(ctx context.Context, e *entity.Product, pg *pagination.Pagination) error {
	var products []*entity.Product

	var conditions []string
	var values []interface{}

	conditions = append(conditions, "is_active = ? AND deleted_at IS NULL")
	values = append(values, true)

	if e.Name != "" {
		conditions = append(conditions, "(name ILIKE ? OR brand ILIKE ? OR description ILIKE ?)")
		search := "%" + e.Name + "%"
		values = append(values, search, search, search)
	}

	if e.CategorySlug != "" {
		conditions = append(conditions, "category_id IN (SELECT id FROM categories WHERE slug = ? AND deleted_at IS NULL)")
		values = append(values, e.CategorySlug)
	} else if e.CategoryID != uuid.Nil {
		conditions = append(conditions, "category_id = ?")
		values = append(values, e.CategoryID)
	}

	query := strings.Join(conditions, " AND ")
	args := append([]interface{}{query}, values...)

	db := r.db.WithContext(ctx).
		Scopes(pagination.Paginate(&entity.Product{}, pg, r.db, args...)).
		Preload("Category").
		Order("is_active DESC, created_at DESC").
		Find(&products)

	if db.Error != nil {
		return handling.NewHttpError500(db.Error)
	}

	pg.Rows = products
	return nil
}

func (r *product) Update(ctx context.Context, e *entity.Product) error {
	tx := r.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()
	defer tx.Rollback()

	result := tx.
		Model(&entity.Product{}).
		Where("id = ? AND deleted_at IS NULL", e.ID).
		Select("name", "slug", "brand", "category_id", "description", "image_url", "image_alt", "price", "is_active").
		Updates(e)
	if result.Error != nil {
		return handling.NewHttpError500(result.Error)
	}
	if result.RowsAffected == 0 {
		return handling.NewHttpError404(nil, "Produk tidak ditemukan atau sudah dihapus.")
	}

	tx.Commit()
	return nil
}

func (r *product) UpdateStatus(ctx context.Context, id uuid.UUID, isActive bool) error {
	tx := r.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()
	defer tx.Rollback()

	result := tx.
		Model(&entity.Product{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Update("is_active", isActive)
	if result.Error != nil {
		return handling.NewHttpError500(result.Error)
	}
	if result.RowsAffected == 0 {
		return handling.NewHttpError404(nil, "Produk tidak ditemukan atau sudah dihapus.")
	}

	tx.Commit()
	return nil
}

func (r *product) Delete(ctx context.Context, id uuid.UUID) error {
	tx := r.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()
	defer tx.Rollback()

	result := tx.
		Where("id = ? AND deleted_at IS NULL", id).
		Delete(&entity.Product{})
	if result.Error != nil {
		return handling.NewHttpError500(result.Error)
	}
	if result.RowsAffected == 0 {
		return handling.NewHttpError404(nil, "Produk tidak ditemukan atau sudah dihapus.")
	}

	tx.Commit()
	return nil
}

func (r *product) SlugExists(ctx context.Context, slug string, excludeID uuid.UUID) (bool, error) {
	var count int64
	db := r.db.WithContext(ctx).Model(&entity.Product{}).
		Where("slug = ? AND deleted_at IS NULL", slug)
	if excludeID != uuid.Nil {
		db = db.Where("id != ?", excludeID)
	}
	result := db.Count(&count)
	if result.Error != nil {
		return false, handling.NewHttpError500(result.Error)
	}
	return count > 0, nil
}

var _ Product = (*product)(nil)
