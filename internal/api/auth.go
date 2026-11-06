package api

import (
	"context"

	"go-memolab-service/internal/core"
	"go-memolab-service/internal/entity"
	"go-memolab-service/internal/value_object/v1/pagination"
)

type Auth interface {
	CheckExistingUserByEmail(ctx context.Context, e *entity.User) error
	CheckExistingUserByPhone(ctx context.Context, e *entity.User) error
	Register(ctx context.Context, e *entity.User, files []*entity.FileUpload, roles []*entity.Role) error

	Login(ctx context.Context, e *entity.Auth) error
	Logout(ctx context.Context, e *entity.User) error

	DeleteAccount(ctx context.Context, e *entity.User) error

	RefreshToken(ctx context.Context, e *entity.Auth) error
	ForgotPassword(ctx context.Context, e *entity.Auth) error
	ResetPassword(ctx context.Context, e *entity.ResetPassword) error
	ChangePassword(ctx context.Context, userID uint, oldPassword, newPassword string) error

	FindUserByID(ctx context.Context, e *entity.User) error

	GetSessions(ctx context.Context, userID uint) ([]*entity.DeviceLog, error)
	UpdateSession(ctx context.Context, e *entity.DeviceLog) error

	UpdateUser(ctx context.Context, e *entity.User) error
	UpdateUserPasswordExpiration(ctx context.Context, e *entity.User) error

	VerifyUserPassword(ctx context.Context, e *entity.User) error

	UpdateProfile(ctx context.Context, e *entity.User) error
	UpdateProfilePhoto(ctx context.Context, e *entity.User) error
	GetActivityLogs(ctx context.Context, pg *pagination.Pagination) error
	GenerateCsrfToken(ctx context.Context, e *entity.Auth) error
	ValidateCsrfToken(ctx context.Context, e *entity.Auth) error

	// FindAllRoles(ctx context.Context, e *entity.UserFindItemAPI) error
	SwitchRole(ctx context.Context, e *entity.Auth) error

	ImpersonateUser(ctx context.Context, e *entity.Auth) error
	ReleaseImpersonation(ctx context.Context, e *entity.Auth) error
}

type auth struct {
	core core.Auth
}

func NewAuth(core core.Auth) Auth {
	return &auth{core: core}
}

func (a *auth) CheckExistingUserByEmail(ctx context.Context, e *entity.User) error {
	return a.core.CheckExistingUserByEmail(ctx, e)
}

func (a *auth) CheckExistingUserByPhone(ctx context.Context, e *entity.User) error {
	return a.core.CheckExistingUserByPhone(ctx, e)
}

func (a *auth) Register(ctx context.Context, e *entity.User, files []*entity.FileUpload, roles []*entity.Role) error {
	return a.core.Register(ctx, e, files, roles)
}

func (a *auth) Login(ctx context.Context, e *entity.Auth) error {
	return a.core.Login(ctx, e)
}

func (a *auth) Logout(ctx context.Context, e *entity.User) error {
	return a.core.Logout(ctx, e)
}

func (a *auth) DeleteAccount(ctx context.Context, e *entity.User) error {
	return nil
}

func (a *auth) RefreshToken(ctx context.Context, e *entity.Auth) error {
	return a.core.RefreshToken(ctx, e)
}

func (a *auth) ForgotPassword(ctx context.Context, e *entity.Auth) error {
	return a.core.ForgotPassword(ctx, e.Username)
}

func (a *auth) ResetPassword(ctx context.Context, e *entity.ResetPassword) error {
	return a.core.ResetPassword(ctx, e.Token, e.Password)
}

func (a *auth) ChangePassword(ctx context.Context, userID uint, oldPassword, newPassword string) error {
	return nil
}

func (a *auth) FindUserByID(ctx context.Context, e *entity.User) error {
	return a.core.FindUserByID(ctx, e)
}

func (a *auth) GetSessions(ctx context.Context, userID uint) ([]*entity.DeviceLog, error) {
	return nil, nil
}

func (a *auth) UpdateSession(ctx context.Context, e *entity.DeviceLog) error {
	return nil
}

func (a *auth) UpdateUser(ctx context.Context, e *entity.User) error {
	return nil
}

func (a *auth) GenerateCsrfToken(ctx context.Context, e *entity.Auth) error {
	return a.core.GenerateCsrfToken(ctx, e)
}

func (a *auth) GetActivityLogs(ctx context.Context, pg *pagination.Pagination) error {
	return nil
}

func (a *auth) ImpersonateUser(ctx context.Context, e *entity.Auth) error {
	return nil
}

func (a *auth) ReleaseImpersonation(ctx context.Context, e *entity.Auth) error {
	return nil
}

func (a *auth) SwitchRole(ctx context.Context, e *entity.Auth) error {
	return nil
}

func (a *auth) UpdateProfile(ctx context.Context, e *entity.User) error {
	return nil
}

func (a *auth) UpdateProfilePhoto(ctx context.Context, e *entity.User) error {
	return nil
}

func (a *auth) UpdateUserPasswordExpiration(ctx context.Context, e *entity.User) error {
	return nil
}

func (a *auth) ValidateCsrfToken(ctx context.Context, e *entity.Auth) error {
	return nil
}

func (a *auth) VerifyUserPassword(ctx context.Context, e *entity.User) error {
	return nil
}
