package api

import (
	"context"
	"go-starter-service/internal/entity"
	"go-starter-service/internal/framework/secondary/repository"
	"go-starter-service/internal/value_object/v1/pagination"

	"github.com/google/uuid"
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

func NewCategory(repo repository.Category) Category {
	return &category{repo: repo}
}

type category struct {
	repo repository.Category
}

func (c *category) Create(ctx context.Context, e *entity.Category) error {
	return c.repo.Create(ctx, e)
}

func (c *category) FindByID(ctx context.Context, id uuid.UUID) (*entity.Category, error) {
	return c.repo.FindByID(ctx, id)
}

func (c *category) FindBySlug(ctx context.Context, slug string) (*entity.Category, error) {
	return c.repo.FindBySlug(ctx, slug)
}

func (c *category) FindAll(ctx context.Context, e *entity.Category, pg *pagination.Pagination) error {
	return c.repo.FindAll(ctx, e, pg)
}

func (c *category) Update(ctx context.Context, e *entity.Category) error {
	return c.repo.Update(ctx, e)
}

func (c *category) Delete(ctx context.Context, id uuid.UUID) error {
	return c.repo.Delete(ctx, id)
}

func (c *category) FindCategories(ctx context.Context) ([]*entity.Category, error) {
	return c.repo.FindCategories(ctx)
}

func (c *category) CategoryExists(ctx context.Context, id uuid.UUID) (bool, error) {
	return c.repo.CategoryExists(ctx, id)
}