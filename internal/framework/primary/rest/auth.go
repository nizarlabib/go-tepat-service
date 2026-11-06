package rest

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"mime"
	"net/http"
	"path/filepath"
	"slices"
	"time"

	"go-memolab-service/internal/api"
	"go-memolab-service/internal/entity"
	"go-memolab-service/internal/framework/secondary/repository"
	"go-memolab-service/internal/framework/secondary/storage"
	"go-memolab-service/internal/middleware"
	"go-memolab-service/internal/pkg/cache"
	"go-memolab-service/internal/pkg/ratelimiter"
	"go-memolab-service/internal/pkg/rbac"
	"go-memolab-service/internal/spesial"
	"go-memolab-service/internal/value_object/v1/constant"
	"go-memolab-service/internal/value_object/v1/convert"
	"go-memolab-service/internal/value_object/v1/handling"
	"go-memolab-service/internal/value_object/v1/pagination"
	"go-memolab-service/internal/value_object/v1/response"
	"go-memolab-service/internal/value_object/v1/router"
	"go-memolab-service/internal/value_object/v3/validator"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var (
	kUserMakerRoleID    = rbac.UUIDPad(2)
	kUserCheckerRoleID  = rbac.UUIDPad(4)
	kUserApproverRoleID = rbac.UUIDPad(6)
	kUserViewerRoleID   = rbac.UUIDPad(8)

	kProviderMakerRoleID    = rbac.UUIDPad(3)
	kProviderCheckerRoleID  = rbac.UUIDPad(5)
	kProviderApproverRoleID = rbac.UUIDPad(7)
	kProviderViewerRoleID   = rbac.UUIDPad(9)
)

type Auth interface {
	CheckExistingUserByEmail(w http.ResponseWriter, r *http.Request) error
	CheckExistingUserByPhone(w http.ResponseWriter, r *http.Request) error
	Register(w http.ResponseWriter, r *http.Request) error

	Login(w http.ResponseWriter, r *http.Request) error
	Logout(w http.ResponseWriter, r *http.Request) error
	DeleteAccount(w http.ResponseWriter, r *http.Request) error

	RefreshToken(w http.ResponseWriter, r *http.Request) error
	GetMe(w http.ResponseWriter, r *http.Request) error
	UpdateProfile(w http.ResponseWriter, r *http.Request) error

	ForgotPassword(w http.ResponseWriter, r *http.Request) error
	ForgotPasswordByPhone(w http.ResponseWriter, r *http.Request) error
	ResetPassword(w http.ResponseWriter, r *http.Request) error
	ChangePassword(w http.ResponseWriter, r *http.Request) error

	GetActivityLogs(w http.ResponseWriter, r *http.Request) error

	// FindAllRoles(w http.ResponseWriter, r *http.Request) error

	ImpersonateUser(w http.ResponseWriter, r *http.Request) error
	ReleaseImpersonation(w http.ResponseWriter, r *http.Request) error

	Start(ctx context.Context) chi.Router
}

func NewAuth(api api.Auth, repoUser repository.User, apiStorage storage.Storage, redis *redis.Client) Auth {
	return &restAuth{
		apiAuth:    api,
		repoUser:   repoUser,
		redis:      redis,
		apiStorage: apiStorage,
	}
}

type restAuth struct {
	apiAuth    api.Auth
	repoUser   repository.User
	redis      *redis.Client
	apiStorage storage.Storage
}

func (c *restAuth) Start(ctx context.Context) chi.Router {
	r := chi.NewRouter()

	r.Method(http.MethodGet, "/csrf_token", router.RootHandler(c.CsrfToken))
	r.Method(http.MethodPost, "/check-existing-user-by-email", router.RootHandler(c.CheckExistingUserByEmail))
	r.Method(http.MethodPost, "/check-existing-user-by-phone", router.RootHandler(c.CheckExistingUserByPhone))

	r.Group(func(r chi.Router) {
		r.Use(ratelimiter.Middleware(10, time.Minute))
		r.Method(http.MethodPost, "/register", router.RootHandler(c.Register))
		r.Method(http.MethodPost, "/login", router.RootHandler(c.Login))
		r.Method(http.MethodPost, "/forgot-password", router.RootHandler(c.ForgotPassword))
		r.Method(http.MethodPost, "/forgot-password/phone", router.RootHandler(c.ForgotPasswordByPhone))
	})

	r.Method(http.MethodPost, "/reset-password", router.RootHandler(c.ResetPassword))

	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(c.repoUser))

		r.Method(http.MethodPost, "/logout", router.RootHandler(c.Logout))
		r.Method(http.MethodGet, "/me", router.RootHandler(c.GetMe))
		r.Method(http.MethodDelete, "/me/delete-account", router.RootHandler(c.DeleteAccount))
		r.Method(http.MethodPut, "/profile", router.RootHandler(c.UpdateProfile))
		r.Method(http.MethodPut, "/profile/photo", router.RootHandler(c.UpdateProfilePhoto))
		r.Method(http.MethodGet, "/profile/download/{filename}", router.RootHandler(c.DownloadFile))

		r.Method(http.MethodPut, "/change-password", router.RootHandler(c.ChangePassword))
		r.Method(http.MethodPost, "/verify-user-password", router.RootHandler(c.VerifyUserPassword))
		r.Method(http.MethodGet, "/activity-logs", router.RootHandler(c.GetActivityLogs))

		// r.Method(http.MethodGet, "/roles", router.RootHandler(c.FindAllRoles))
		r.Method(http.MethodPost, "/switch_role", router.RootHandler(c.SwitchRole))

		r.With(rbac.WithPermission(rbac.MustSuperAdmin())).
			Method(http.MethodPost, "/impersonate/{user_id}", router.RootHandler(c.ImpersonateUser))
		r.Method(http.MethodPost, "/release_impersonation/{user_id}", router.RootHandler(c.ReleaseImpersonation))
	})

	// update profile photo request

	r.With(middleware.RefreshTokenMiddleware(c.repoUser)).
		Method(http.MethodPost, "/refresh-token", router.RootHandler(c.RefreshToken))

	return r
}

// ImpersonateUser implements [Auth].
func (c *restAuth) ImpersonateUser(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	userID := convert.IDUint(r, "user_id")
	if userID == 0 {
		return handling.NewHttpError(nil, http.StatusBadRequest, "user_id wajib diisi.", constant.ERR_VALIDATION_ERROR)
	}

	auth := new(entity.Auth)
	auth.UserID = userID

	if err := c.apiAuth.ImpersonateUser(ctx, auth); err != nil {
		return err
	}

	response.SuccessLogin("impersonate user success", auth.AccessToken, auth.RefreshToken, w)
	return nil
}

// ReleaseImpersonation implements [Auth].
func (c *restAuth) ReleaseImpersonation(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	userID := convert.IDUint(r, "user_id")
	if userID == 0 {
		return handling.NewHttpError(nil, http.StatusBadRequest, "user_id wajib diisi.", constant.ERR_VALIDATION_ERROR)
	}

	auth := new(entity.Auth)
	auth.UserID = userID

	if err := c.apiAuth.ReleaseImpersonation(ctx, auth); err != nil {
		return err
	}

	response.SuccessLogin("release impersonation success", auth.AccessToken, auth.RefreshToken, w)
	return nil
}

func (c *restAuth) GetActivityLogs(w http.ResponseWriter, r *http.Request) error {
	pg := pagination.NewPagination(r.URL.Query())

	if err := c.apiAuth.GetActivityLogs(r.Context(), pg); err != nil {
		return err
	}

	response.SuccessDataPaginate("sukses mengambil data log aktivitas", "activity_logs", pg.Rows, &response.Page{
		TotalData:   pg.TotalRows,
		Limit:       pg.GetLimit(),
		CurrentPage: pg.GetPage(),
		TotalPage:   pg.TotalPages,
	}, w)
	return nil
}

func (c *restAuth) DeleteAccount(w http.ResponseWriter, r *http.Request) error {
	if err := validateCsrf(r, c); err != nil {
		return err
	}

	userID, err := spesial.GetUserIDFromContext(r.Context())
	if err != nil {
		return err
	}
	req := deleteAccountForm{}
	if err = validator.DecodeValidate(r.Body, &req); err != nil {
		return err
	}
	e := req.toEntity()
	e.ID = userID
	if err = c.apiAuth.DeleteAccount(r.Context(), e); err != nil {
		return err
	}

	response.Success("akun berhasil dihapus", w)
	return nil
}

func (c *restAuth) Register(w http.ResponseWriter, r *http.Request) error {
	if err := validateCsrf(r, c); err != nil {
		return err
	}

	var err error

	if r.MultipartForm == nil {
		err = r.ParseMultipartForm(32 << 20)
		if err != nil {
			return handling.NewHttpError(err, http.StatusBadRequest, "Failed to parse multipart form.", constant.ERR_VALIDATION_ERROR)
		}
	}

	if r.MultipartForm == nil || r.MultipartForm.File == nil {
		return handling.NewHttpError(nil, http.StatusBadRequest, "Invalid request.", constant.ERR_VALIDATION_ERROR)
	}

	req := registerForm{
		Fullname:       r.FormValue("fullname"),
		Phone:          r.FormValue("phone"),
		Email:          r.FormValue("email"),
		Password:       r.FormValue("password"),
		Gender:         r.FormValue("gender"),
		RoleID:         r.FormValue("role_id"),
		ProfilePicture: r.FormValue("profile_picture"),
	}

	if err = validator.Validate(req); err != nil {
		return err
	}

	e := req.toEntity()

	fileValidator := validator.NewFileValidator(r.MultipartForm)

	// imageExts := []string{"png", "jpeg", "jpg"}
	// pdfExts := []string{"pdf"}
	// videoExts := []string{"webm"}

	// e.FileVideoLiveness, err = fileValidator.Validate("file_video_liveness", videoExts, 5)
	// if err != nil {
	// 	return err
	// }

	// the following roles are fixed template
	var userRoles []*entity.Role

	if err := c.apiAuth.Register(r.Context(), e, fileValidator.Files(), userRoles); err != nil {
		return err
	}

	response.Success("user successfuly registered", w)
	return nil
}

func validateCsrf(r *http.Request, c *restAuth) error {
	xCsrfToken := r.Header.Get("X-CSRF-Token")

	if xCsrfToken == "" {
		return handling.NewHttpError(nil, http.StatusBadRequest, "CSRF token is required.", constant.ERR_RECORD_NOT_FOUND)
	}

	ea := new(entity.Auth)
	ea.XCsrfToken = xCsrfToken
	if err := c.apiAuth.ValidateCsrfToken(r.Context(), ea); err != nil {
		return err
	}

	return nil
}

func (c *restAuth) Login(w http.ResponseWriter, r *http.Request) error {
	if err := validateCsrf(r, c); err != nil {
		return err
	}

	req := loginForm{}
	if err := validator.DecodeValidate(r.Body, &req); err != nil {
		return err
	}
	e := req.toEntity()

	err := c.apiAuth.Login(r.Context(), e)
	if err != nil {
		return err
	}

	response.SuccessLogin("Login Sucessfull", e.AccessToken, e.RefreshToken, w)
	return nil
}

func (c *restAuth) Logout(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	authClaims, err := spesial.GetAuthClaimsFromContext(r.Context())
	if err != nil {
		return err
	}

	if authClaims.IsImpersonated() {
		auth := new(entity.Auth)
		auth.UserID = authClaims.UserID
		if err := c.apiAuth.ReleaseImpersonation(ctx, auth); err != nil {
			return err
		}
		response.Success("success release impersonation", w)
		return nil
	}

	userID, err := spesial.GetUserIDFromContext(ctx)
	if err != nil {
		return err
	}
	if err := c.apiAuth.Logout(ctx, entity.NewUserFromID(userID)); err != nil {
		return err
	}

	if err := c.repoUser.BlockToken(ctx, authClaims.JTI, int(time.Until(authClaims.ExpiresAt.Time).Seconds())); err != nil {
		slog.Warn("failed to block token on logout", slog.String("jti", authClaims.JTI), slog.String("error", err.Error()))
	}
	response.Success("success logout", w)
	return nil
}

// ForgotPasswordByPhone implements [Auth].
func (c *restAuth) ForgotPasswordByPhone(w http.ResponseWriter, r *http.Request) error {
	if err := validateCsrf(r, c); err != nil {
		return err
	}
	req := forgotPasswordByPhoneForm{}
	if err := validator.DecodeValidate(r.Body, &req); err != nil {
		return err
	}
	e := req.toEntity()
	if err := c.apiAuth.ForgotPassword(r.Context(), e); err != nil {
		return err
	}
	response.Success("whatsapp has been sent.", w)
	return nil
}

func (c *restAuth) ForgotPassword(w http.ResponseWriter, r *http.Request) error {
	if err := validateCsrf(r, c); err != nil {
		return err
	}
	req := forgotPasswordForm{}
	if err := validator.DecodeValidate(r.Body, &req); err != nil {
		return err
	}
	e := req.toEntity()
	if err := c.apiAuth.ForgotPassword(r.Context(), e); err != nil {
		return err
	}
	response.Success("email has been sent.", w)
	return nil
}

func (c *restAuth) ResetPassword(w http.ResponseWriter, r *http.Request) error {
	if err := validateCsrf(r, c); err != nil {
		return err
	}

	req := resetPasswordForm{}
	if err := validator.DecodeValidate(r.Body, &req); err != nil {
		return err
	}
	e := req.toEntity()
	if err := c.apiAuth.ResetPassword(r.Context(), e); err != nil {
		return err
	}
	response.Success("password berhasil di reset", w)
	return nil
}

type userMe struct {
	*entity.User

	RoleID         uuid.UUID `json:"role_id"`
	Role           string    `json:"role"`
	Permissions    string    `json:"permissions"`
	PermissionMask string    `json:"permission_mask"`

	ImpersonatedBy uint `json:"impersonated_by,omitempty"`
}

func (c *restAuth) GetMe(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	user, err := spesial.GetUserFromContext(ctx)
	if err != nil {
		return err
	}

	authClaims, err := spesial.GetAuthClaimsFromContext(ctx)
	if err != nil {
		return err
	}

	perms, err := spesial.GetBigPermsFromContext(ctx)
	if err != nil {
		log.Println(err)
		return err
	}

	if authClaims.RoleID == uuid.Nil {
		if user.Role != nil {
			authClaims.RoleID = user.Role.ID
			authClaims.Role = user.Role.Name
		}
	}

	userme := &userMe{
		User:           user,
		RoleID:         authClaims.RoleID,
		Role:           authClaims.Role,
		Permissions:    perms.String(),
		PermissionMask: perms.String(),

		ImpersonatedBy: authClaims.ImpersonatedBy,
	}

	response.SuccessData("success getting me", "me", userme, w)
	return nil
}

func (c *restAuth) CheckExistingUserByEmail(w http.ResponseWriter, r *http.Request) error {
	var err error
	req := checkExistingUserByEmailForm{}

	if err = validator.DecodeValidate(r.Body, &req); err != nil {
		return err
	}

	e := req.toEntity()

	if err = c.apiAuth.CheckExistingUserByEmail(r.Context(), e); err != nil {
		return err
	}

	response.Success("email belum terdaftar", w)
	return nil
}

func (c *restAuth) CheckExistingUserByPhone(w http.ResponseWriter, r *http.Request) error {
	var err error
	req := checkExistingUserByPhoneForm{}

	if err = validator.DecodeValidate(r.Body, &req); err != nil {
		return err
	}

	e := req.toEntity()

	if err = c.apiAuth.CheckExistingUserByPhone(r.Context(), e); err != nil {
		return err
	}

	response.Success("phone belum terdaftar", w)
	return nil
}

func (c *restAuth) VerifyUserPassword(w http.ResponseWriter, r *http.Request) error {
	userID, err := spesial.GetUserIDFromContext(r.Context())
	if err != nil {
		return err
	}

	req := verifyUserPasswordForm{}

	if err = validator.DecodeValidate(r.Body, &req); err != nil {
		return err
	}

	e := &entity.User{
		BaseModel: entity.BaseModel{
			ID: userID,
		},
		Password: req.Password,
	}

	if err = c.apiAuth.VerifyUserPassword(r.Context(), e); err != nil {
		return err
	}

	response.Success("password benar", w)
	return nil
}

func (c *restAuth) UpdateProfile(w http.ResponseWriter, r *http.Request) error {
	if err := validateCsrf(r, c); err != nil {
		return err
	}

	ctx := r.Context()

	userID, err := spesial.GetUserIDFromContext(r.Context())
	if err != nil {
		return err
	}

	req := updateProfileForm{}
	if err = validator.Decode(r.Body, &req); err != nil {
		return err
	}

	e := req.toEntity()
	e.ID = userID

	if err = c.apiAuth.UpdateProfile(ctx, e); err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("user_me:%s", convert.UintToStr(e.ID))
	cacheKey2 := fmt.Sprintf("me:%s", convert.UintToStr(e.ID))
	cacheKey3 := fmt.Sprintf("user:%s", convert.UintToStr(e.ID))
	_ = cache.RevalidateKeys(ctx, c.redis, cacheKey, cacheKey2, cacheKey3)

	response.Success("profile berhasil diperbarui", w)
	return nil
}

func (c *restAuth) RefreshToken(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	authClaims, err := spesial.GetAuthClaimsFromContext(ctx)
	if err != nil {
		return err
	}
	userID, err := spesial.GetUserIDFromContext(ctx)
	if err != nil {
		return err
	}
	refreshToken, err := middleware.GetRefreshTokenFromContext(ctx)
	if err != nil {
		return err
	}

	if err := c.repoUser.BlockToken(ctx, authClaims.JTI, int(time.Until(authClaims.ExpiresAt.Time).Seconds())); err != nil {
		slog.Warn("failed to block old refresh token", slog.String("jti", authClaims.JTI), slog.String("error", err.Error()))
	}

	e := &entity.Auth{
		UserID:       userID,
		SessionID:    authClaims.SessionID,
		RefreshToken: refreshToken,
	}

	if err := c.apiAuth.RefreshToken(ctx, e); err != nil {
		return err
	}

	response.SuccessLogin("refresh token sucessfull", e.AccessToken, e.RefreshToken, w)
	return nil
}

func (c *restAuth) ChangePassword(w http.ResponseWriter, r *http.Request) error {
	if err := validateCsrf(r, c); err != nil {
		return err
	}

	userID, err := spesial.GetUserIDFromContext(r.Context())
	if err != nil {
		return err
	}
	req := changePasswordForm{}
	if err := validator.DecodeValidate(r.Body, &req); err != nil {
		return err
	}
	if err := c.apiAuth.ChangePassword(r.Context(), userID, req.OldPassword, req.NewPassword); err != nil {
		return err
	}
	response.Success("password berhasil di ubah", w)
	return nil
}

func (c *restAuth) UpdateProfilePhoto(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	userID, err := spesial.GetUserIDFromContext(r.Context())
	if err != nil {
		return err
	}

	req := updateProfilePhoto{}
	if err := validator.DecodeValidate(r.Body, &req); err != nil {
		return err
	}

	e := req.toEntity()
	e.ID = userID

	if err = c.apiAuth.UpdateProfilePhoto(ctx, e); err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("profile_pic:%s", convert.UintToStr(e.ID))
	_ = cache.RevalidateKeys(ctx, c.redis, cacheKey)

	response.Success("update profile photo berhasil", w)
	return nil
}

func (c *restAuth) CsrfToken(w http.ResponseWriter, r *http.Request) error {
	e := new(entity.Auth)
	if err := c.apiAuth.GenerateCsrfToken(r.Context(), e); err != nil {
		return err
	}
	response.SuccessData("CSRF token generated", "csrf_token", e.CsrfToken, w)
	return nil
}

func (c *restAuth) GetProfilePhoto(w http.ResponseWriter, r *http.Request) error {
	userID, err := spesial.GetUserIDFromContext(r.Context())
	if err != nil {
		return err
	}

	e := &entity.User{BaseModel: entity.BaseModel{ID: userID}}
	if err = c.apiAuth.FindUserByID(r.Context(), e); err != nil {
		return err
	}

	if e.ProfilePictureURL == nil {
		response.Success("photo kosong", w)
		return nil
	}

	filename := fmt.Sprintf("%s/%s", "registration", *e.ProfilePictureURL)
	buff, err := c.apiStorage.GetFile(r.Context(), filename)
	if err != nil {
		return err
	}

	mimeType := mime.TypeByExtension(filepath.Ext(*e.ProfilePictureURL))
	w.Header().Set("Content-Type", mimeType)

	_, err = w.Write(buff)
	if err != nil {
		return handling.NewHttpError500(err)
	}

	return nil
}

func (c *restAuth) DownloadFile(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	userID, err := spesial.GetUserIDFromContext(ctx)
	if err != nil {
		return err
	}

	e := &entity.User{BaseModel: entity.BaseModel{ID: userID}}
	if err = c.apiAuth.FindUserByID(ctx, e); err != nil {
		return err
	}

	filenames := []string{}
	if e.ProfilePictureURL != nil {
		filenames = append(filenames, *e.ProfilePictureURL)
	}

	filename := r.PathValue("filename")
	if err = validator.IsValidFilename(filename); err != nil {
		return err
	}

	if !slices.Contains(filenames, filename) {
		return handling.NewHttpError403(nil)
	}

	isProfilePicture := e.ProfilePictureURL != nil && filename == *e.ProfilePictureURL
	cacheKey := fmt.Sprintf("profile_pic:%s", convert.UintToStr(e.ID))
	if isProfilePicture {
		encoded, err := c.redis.Get(ctx, cacheKey).Result()
		if err == nil {
			decoded, err := convert.Base64Decode(encoded)
			if err == nil {
				w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
				w.Header().Set("Content-Type", "application/octet-stream")
				_, err = w.Write(decoded)
				if err != nil {
					return handling.NewHttpError404(err, fmt.Sprintf("Berkas %s tidak ditemukan.", filename))
				}
				return nil
			} else {
				_ = c.redis.Expire(ctx, cacheKey, 2*time.Second)
			}
		}
	}

	filenameObject := fmt.Sprintf("%s/%s", "registration", filename)

	buff, err := c.apiStorage.GetFile(ctx, filenameObject)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Header().Set("Content-Type", "application/octet-stream")

	_, err = w.Write(buff)
	if err != nil {
		return handling.NewHttpError404(err, fmt.Sprintf("Berkas %s tidak ditemukan.", filename))
	}

	if isProfilePicture {
		buffB64 := convert.Base64Encode(buff)
		_ = c.redis.Set(ctx, cacheKey, buffB64, 1*time.Hour)
	}

	return nil
}

// func (c *restAuth) FindAllRoles(w http.ResponseWriter, r *http.Request) error {
// 	userID, err := middleware.GetUserIDFromContext(r.Context())
// 	if err != nil {
// 		return err
// 	}

// 	user := new(entity.User)
// 	user.ID = userID
// 	e := &entity.UserFindItemAPI{
// 		User: user,
// 	}

// 	err = c.apiAuth.FindAllRoles(r.Context(), e)
// 	if err != nil {
// 		return err
// 	}

// 	response.SuccessData("sukses mengambil data roles", "roles_me", e.Roles, w)
// 	return nil
// }

func (c *restAuth) SwitchRole(w http.ResponseWriter, r *http.Request) error {
	if err := validateCsrf(r, c); err != nil {
		return err
	}

	userID, err := spesial.GetUserIDFromContext(r.Context())
	if err != nil {
		return err
	}

	req := switchRoleForm{}
	if err = validator.Decode(r.Body, &req); err != nil {
		return err
	}

	roleID, err := uuid.Parse(req.RoleID)
	if err != nil {
		return handling.NewHttpError(err, http.StatusBadRequest, "role_id bukan UUID yang valid.", constant.ERR_VALIDATION_ERROR)
	}

	e := new(entity.Auth)
	e.UserID = userID
	e.RoleID = roleID

	err = c.apiAuth.SwitchRole(r.Context(), e)
	if err != nil {
		return err
	}

	response.SuccessLogin("sukses switch role", e.AccessToken, e.RefreshToken, w)
	return nil
}
