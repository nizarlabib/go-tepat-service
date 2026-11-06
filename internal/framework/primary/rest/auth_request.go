package rest

import (
	"strings"

	"go-starter-service/internal/entity"

	"github.com/google/uuid"
)

type updateProfilePhoto struct {
	ProfilePictureUrl string `json:"profile_picture_url" validate:"required"`
}

func (d *updateProfilePhoto) toEntity() *entity.User {
	return &entity.User{
		ProfilePictureURL: &d.ProfilePictureUrl,
	}
}

type updateProfileForm struct {
	Fullname string `json:"fullname" `
	Gender   string `json:"gender" `
	Email    string `json:"email" `
	Phone    string `json:"phone"`
}

// validate:"required"
func (d *updateProfileForm) toEntity() *entity.User {
	return &entity.User{
		Fullname: d.Fullname,
		Email:    d.Email,
		Phone:    d.Phone,
		Gender:   entity.Gender(d.Gender),
	}
}

type registerForm struct {
	Fullname       string `json:"fullname" validate:"required"`
	Email          string `json:"email" validate:"required,email"`
	Phone          string `json:"phone" validate:"required,phone"`
	Password       string `json:"password" validate:"required,password"`
	Gender         string `json:"gender" validate:"required,oneof=male female"`
	ProfilePicture string `json:"profile_picture"`
	RoleID         string `json:"role_id" validate:"required"`
}

func (m registerForm) toEntity() *entity.User {
	roleID := uuid.MustParse(m.RoleID)
	return &entity.User{
		Fullname:          m.Fullname,
		Email:             strings.ToLower(strings.TrimSpace(m.Email)),
		Phone:             m.Phone,
		Password:          m.Password,
		AccountStatus:     entity.UserAccountStatusPending,
		Gender:            entity.Gender(m.Gender),
		ProfilePictureURL: &m.ProfilePicture,
		RoleID:            &roleID,
	}
}

type loginForm struct {
	// deprecated
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password" validate:"required"`

	Latitude  string `json:"latitude,omitempty" validate:"omitempty,latitude"`
	Longitude string `json:"longitude,omitempty" validate:"omitempty,longitude"`
}

func (d loginForm) toEntity() *entity.Auth {
	return &entity.Auth{
		Username: strings.TrimSpace(d.Username),
		Email:    strings.ToLower(strings.TrimSpace(d.Email)),
		Password: strings.TrimSpace(d.Password),

		Latitude:  d.Latitude,
		Longitude: d.Longitude,
	}
}

type forgotPasswordByPhoneForm struct {
	Phone string `json:"phone" validate:"required,phone"`
}

func (d forgotPasswordByPhoneForm) toEntity() *entity.Auth {
	return &entity.Auth{
		Username: strings.ToLower(strings.TrimSpace(d.Phone)),
	}
}

type forgotPasswordForm struct {
	Email string `json:"email" validate:"required,email"`
}

func (d forgotPasswordForm) toEntity() *entity.Auth {
	return &entity.Auth{
		Username: strings.ToLower(strings.TrimSpace(d.Email)),
	}
}

type resetPasswordForm struct {
	Token    string `json:"token" validate:"required"`
	Password string `json:"password" validate:"required,password"`
}

func (d resetPasswordForm) toEntity() *entity.ResetPassword {
	return &entity.ResetPassword{
		Token:    d.Token,
		Password: d.Password,
	}
}

type verifyUserPasswordForm struct {
	Password string `json:"password" validate:"required"`
}

type checkExistingUserByEmailForm struct {
	Email string `json:"email" validate:"required,email"`
}

func (d checkExistingUserByEmailForm) toEntity() *entity.User {
	return &entity.User{
		Email: strings.ToLower(strings.TrimSpace(d.Email)),
	}
}

type checkExistingUserByPhoneForm struct {
	Phone string `json:"phone" validate:"required,phone"`
}

func (d checkExistingUserByPhoneForm) toEntity() *entity.User {
	return &entity.User{
		Phone: d.Phone,
	}
}

type changePasswordForm struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,password"`
}

type deleteAccountForm struct {
	Password string  `json:"password" validate:"required"`
	Reason   *string `json:"reason,omitempty"`
}

func (e *deleteAccountForm) toEntity() *entity.User {
	return &entity.User{
		AccountDeletionReason: e.Reason,
		Password:              e.Password,
	}
}

type generateIntegrationApiKeyTokenForm struct {
	AccessKey string `json:"access_key" validate:"required"`
	SecretKey string `json:"secret_key" validate:"required"`
}

type checkExistingUserByNikForm struct {
	Nik string `json:"nik" validate:"required,len=16"`
}

type switchRoleForm struct {
	RoleID string `json:"role_id" validate:"required"`
}
