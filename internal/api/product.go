package api

import (
	"context"

	"go-starter-service/internal/core"
	"go-starter-service/internal/entity"
	"go-starter-service/internal/value_object/v1/pagination"

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

type product struct {
	core core.Product
}

func NewProduct(core core.Product) Product {
	return &product{core: core}
}

func (p *product) Create(ctx context.Context, e *entity.Product) error {
	return p.core.Create(ctx, e)
}

func (p *product) FindByID(ctx context.Context, id uuid.UUID) (*entity.Product, error) {
	return p.core.FindByID(ctx, id)
}

func (p *product) FindBySlug(ctx context.Context, slug string) (*entity.Product, error) {
	return p.core.FindBySlug(ctx, slug)
}

func (p *product) FindAll(ctx context.Context, e *entity.Product, pg *pagination.Pagination) error {
	return p.core.FindAll(ctx, e, pg)
}

func (p *product) Update(ctx context.Context, id uuid.UUID, e *entity.Product) error {
	return p.core.Update(ctx, id, e)
}

func (p *product) UpdateStatus(ctx context.Context, id uuid.UUID, isActive bool) error {
	return p.core.UpdateStatus(ctx, id, isActive)
}

func (p *product) Delete(ctx context.Context, id uuid.UUID) error {
	return p.core.Delete(ctx, id)
}

var _ Product = (*product)(nil)
