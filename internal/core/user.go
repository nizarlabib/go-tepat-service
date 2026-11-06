package core

import (
	"context"

	"go-starter-service/internal/entity"
	"go-starter-service/internal/framework/secondary/repository"
	"go-starter-service/internal/value_object/v1/pagination"
)

type User interface {
	FindAll(ctx context.Context, pg *pagination.Pagination, opts entity.UserFindAllOptions) error
	FindByPhone(ctx context.Context, e *entity.User) error
	FindByEmail(ctx context.Context, e *entity.User) error
	FindByIDWithRoles(ctx context.Context, e *entity.UserFindItemAPI) error
	FindByID(ctx context.Context, e *entity.User) error
	FindByIDs(ctx context.Context, ids []uint) ([]*entity.UserFindItemAPI, error)
}

type user struct {
	repo repository.User
}

func NewUser(repo repository.User) User {
	return &user{repo: repo}
}

func (u *user) FindAll(ctx context.Context, pg *pagination.Pagination, opts entity.UserFindAllOptions) error {
	return u.repo.FindAll(ctx, pg, opts)
}

func (u *user) FindByPhone(ctx context.Context, e *entity.User) error {
	return u.repo.FindByPhone(ctx, e)
}

func (u *user) FindByEmail(ctx context.Context, e *entity.User) error {
	return u.repo.FindByEmail(ctx, e)
}

func (u *user) FindByIDWithRoles(ctx context.Context, e *entity.UserFindItemAPI) error {
	return u.repo.FindByIDWithRoles(ctx, e)
}

func (u *user) FindByIDs(ctx context.Context, ids []uint) ([]*entity.UserFindItemAPI, error) {
	return u.repo.FindByIDs(ctx, ids)
}

func (u *user) FindByID(ctx context.Context, e *entity.User) error {
	return u.repo.FindByID(ctx, e)
}
