package core

import (
	"context"
	"fmt"

	"go-starter-service/internal/entity"
	"go-starter-service/internal/framework/secondary/repository"
	"go-starter-service/internal/value_object/v1/handling"
	"go-starter-service/internal/value_object/v1/pagination"
	utilutils "go-starter-service/internal/value_object/v1/utils"

	"github.com/google/uuid"
)

type Product interface {
	Create(ctx context.Context, e *entity.Product) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Product, error)
	FindBySlug(ctx context.Context, slug string) (*entity.Product, error)
	FindAll(ctx context.Context, e *entity.Product, pg *pagination.Pagination) error
	Update(ctx context.Context, id uuid.UUID, e *entity.Product) error
	UpdateStatus(ctx context.Context, id uuid.UUID, isActive bool) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type productOptions func(fac *product)

func NewProduct(opts ...productOptions) Product {
	fac := &product{}
	for _, opt := range opts {
		opt(fac)
	}
	return fac
}

func ProductWithRepo(repo repository.Product) productOptions {
	return func(fac *product) {
		fac.repo = repo
	}
}

func ProductWithCategory(category Category) productOptions {
	return func(fac *product) {
		fac.category = category
	}
}

type product struct {
	repo     repository.Product
	category Category
}

func (p *product) Create(ctx context.Context, e *entity.Product) error {
	if e.Name == "" {
		return handling.NewHttpError(nil, 400, "Nama produk wajib diisi.", "ERR_PRODUCT_NAME_REQUIRED")
	}

	if e.CategoryID == uuid.Nil {
		return handling.NewHttpError(nil, 400, "Category ID wajib diisi.", "ERR_PRODUCT_CATEGORY_REQUIRED")
	}

	exists, err := p.category.CategoryExists(ctx, e.CategoryID)
	if err != nil {
		return err
	}
	if !exists {
		return handling.NewHttpError(nil, 400, "Kategori tidak valid atau tidak ditemukan.", "ERR_PRODUCT_CATEGORY_NOT_FOUND")
	}

	if e.Price.IsNull() == false {
		if e.Price.Float64() < 0 {
			return handling.NewHttpError(nil, 400, "Harga tidak boleh negatif.", "ERR_PRODUCT_PRICE_NEGATIVE")
		}
	}

	slug := e.Slug
	if slug == "" {
		slug = utilutils.Slugify(e.Name)
	}
	slug, err = p.ensureUniqueSlug(ctx, slug, uuid.Nil)
	if err != nil {
		return err
	}
	e.Slug = slug

	// Ignore client-provided ID and timestamps.
	e.ID = uuid.Nil

	return p.repo.Create(ctx, e)
}

func (p *product) FindByID(ctx context.Context, id uuid.UUID) (*entity.Product, error) {
	return p.repo.FindByID(ctx, id)
}

func (p *product) FindBySlug(ctx context.Context, slug string) (*entity.Product, error) {
	return p.repo.FindBySlug(ctx, slug)
}

func (p *product) FindAll(ctx context.Context, e *entity.Product, pg *pagination.Pagination) error {
	return p.repo.FindAll(ctx, e, pg)
}

func (p *product) Update(ctx context.Context, id uuid.UUID, e *entity.Product) error {
	if e.Name == "" {
		return handling.NewHttpError(nil, 400, "Nama produk wajib diisi.", "ERR_PRODUCT_NAME_REQUIRED")
	}

	if e.CategoryID == uuid.Nil {
		return handling.NewHttpError(nil, 400, "Category ID wajib diisi.", "ERR_PRODUCT_CATEGORY_REQUIRED")
	}

	exists, err := p.category.CategoryExists(ctx, e.CategoryID)
	if err != nil {
		return err
	}
	if !exists {
		return handling.NewHttpError(nil, 400, "Kategori tidak valid atau tidak ditemukan.", "ERR_PRODUCT_CATEGORY_NOT_FOUND")
	}

	if e.Price.IsNull() == false {
		if e.Price.Float64() < 0 {
			return handling.NewHttpError(nil, 400, "Harga tidak boleh negatif.", "ERR_PRODUCT_PRICE_NEGATIVE")
		}
	}

	slug := e.Slug
	if slug == "" {
		slug = utilutils.Slugify(e.Name)
	}
	slug, err = p.ensureUniqueSlug(ctx, slug, id)
	if err != nil {
		return err
	}
	e.Slug = slug

	// Do not change ID or timestamps.
	e.ID = id

	return p.repo.Update(ctx, e)
}

func (p *product) UpdateStatus(ctx context.Context, id uuid.UUID, isActive bool) error {
	return p.repo.UpdateStatus(ctx, id, isActive)
}

func (p *product) Delete(ctx context.Context, id uuid.UUID) error {
	return p.repo.Delete(ctx, id)
}

func (p *product) ensureUniqueSlug(ctx context.Context, slug string, excludeID uuid.UUID) (string, error) {
	exists, err := p.repo.SlugExists(ctx, slug, excludeID)
	if err != nil {
		return "", err
	}
	if !exists {
		return slug, nil
	}

	for i := 1; i <= 100; i++ {
		candidate := fmt.Sprintf("%s-%d", slug, i)
		exists, err := p.repo.SlugExists(ctx, candidate, excludeID)
		if err != nil {
			return "", err
		}
		if !exists {
			return candidate, nil
		}
	}

	return "", handling.NewHttpError(nil, 400, "Gagal menghasilkan slug unik.", "ERR_PRODUCT_SLUG_NOT_UNIQUE")
}

var _ Product = (*product)(nil)
