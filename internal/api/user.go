package api

import (
	"context"

	"go-starter-service/internal/core"
	"go-starter-service/internal/entity"
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
	core core.User
}

func NewUser(core core.User) User {
	return &user{core: core}
}

func (u *user) FindAll(ctx context.Context, pg *pagination.Pagination, opts entity.UserFindAllOptions) error {
	return u.core.FindAll(ctx, pg, opts)
}

func (u *user) FindByPhone(ctx context.Context, e *entity.User) error {
	return u.core.FindByPhone(ctx, e)
}

func (u *user) FindByEmail(ctx context.Context, e *entity.User) error {
	return u.core.FindByEmail(ctx, e)
}

func (u *user) FindByIDWithRoles(ctx context.Context, e *entity.UserFindItemAPI) error {
	return u.core.FindByIDWithRoles(ctx, e)
}

func (u *user) FindByIDs(ctx context.Context, ids []uint) ([]*entity.UserFindItemAPI, error) {
	return u.core.FindByIDs(ctx, ids)
}

func (u *user) FindByID(ctx context.Context, e *entity.User) error {
	return u.core.FindByID(ctx, e)
}
