package rest

import (
	"context"
	"go-starter-service/internal/api"
	"go-starter-service/internal/entity"
	"go-starter-service/internal/pkg/rbac"
	permissionindex "go-starter-service/internal/pkg/rbac/permission-index"
	"go-starter-service/internal/value_object/v1/constant"
	"go-starter-service/internal/value_object/v1/handling"
	"go-starter-service/internal/value_object/v1/pagination"
	"go-starter-service/internal/value_object/v1/response"
	"go-starter-service/internal/value_object/v1/router"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type User interface {
	FindAll(w http.ResponseWriter, r *http.Request) error
	// FindByUserID(w http.ResponseWriter, r *http.Request) error
	// FindUsersByIDs(w http.ResponseWriter, r *http.Request) error

	Start(ctx context.Context) chi.Router
}

func NewUser(api api.User) User {
	return &restUser{
		apiUser: api,
	}
}

type restUser struct {
	apiUser api.User
}

func (c *restUser) Start(ctx context.Context) chi.Router {
	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(rbac.WithPermission(rbac.MustHasAll(permissionindex.ReadUser)))

			r.Method(http.MethodGet, "/", router.RootHandler(c.FindAll))
			// r.Method(http.MethodGet, "/count_need_action", router.RootHandler(c.CountNeedAction))
			// r.Method(http.MethodGet, "/{id}", router.RootHandler(c.FindByUserID))
			// r.Method(http.MethodGet, "/{id}/download/{filename}", router.RootHandler(c.DownloadFile))
			// r.Method(http.MethodPost, "/ids", router.RootHandler(c.FindUsersByIDs))
		})

		// r.Method(http.MethodGet, "/sessions", router.RootHandler(c.GetSessions))
		// r.Method("DELETE", "/sessions/{id}", router.RootHandler(c.DeleteSession))

		// newDigitalSignRoutes(r, c)

		// r.With(rbac.WithPermission(rbac.MustHasAll(permissionindex.DeleteUser))).
		// 	Method(http.MethodDelete, "/{id}", router.RootHandler(c.DeleteAccount))

		// r.With(rbac.WithPermission(rbac.MustHasAll(permissionindex.CreateUser))).
		// 	Method(http.MethodPost, "/{id}/restore", router.RootHandler(c.RestoreAccount))

		// r.Group(func(r chi.Router) {
		// 	r.Use(rbac.WithPermission(rbac.MustHasAll(permissionindex.UpdateUser)))

		// 	r.Method(http.MethodPut, "/{id}/reject", router.RootHandler(c.RejectRegistration))
		// 	r.Method(http.MethodPut, "/{id}/approve", router.RootHandler(c.ApproveRegistration))
		// 	r.Method(http.MethodPut, "/{id}/suspend", router.RootHandler(c.Suspend))
		// 	r.Method(http.MethodPut, "/{id}/re_activate", router.RootHandler(c.ReActivate))
		// 	r.Method(http.MethodPut, "/{id}", router.RootHandler(c.UpdateUser))
		// })
	})

	return r
}

// func (c *restUser) CountNeedAction(w http.ResponseWriter, r *http.Request) error {
// 	count, err := c.apiUser.CountNeedAction(r.Context())
// 	if err != nil {
// 		return err
// 	}
// 	response.SuccessData("count need action sukses", "pending_user_count", count, w)
// 	return nil
// }

// func (c *restUser) UpdateUser(w http.ResponseWriter, r *http.Request) error {
// 	userID := convert.IDUint(r)
// 	if userID == 0 {
// 		return handling.NewHttpError404(nil, "Pengguna tidak ditemukan.")
// 	}

// 	var req updateUserForm
// 	if err := validator.DecodeValidate(r.Body, &req); err != nil {
// 		return err
// 	}

// 	e := req.toEntity()
// 	e.ID = userID
// 	var newRoles []*entity.Role
// 	for _, roleIDStr := range req.RoleIds {
// 		roleID, err := uuid.Parse(roleIDStr)
// 		if err != nil {
// 			return handling.NewHttpError(nil, http.StatusBadRequest, fmt.Sprintf("role_id '%s' tidak valid", roleIDStr), constant.ERR_VALIDATION_ERROR)
// 		}
// 		newRoles = append(newRoles, &entity.Role{ID: roleID})
// 	}

// 	if err := c.apiUser.UpdateUser(r.Context(), e, newRoles); err != nil {
// 		return err
// 	}

// 	response.SuccessData("user berhasil diupdate", "id", userID, w)
// 	return nil
// }

// func (c *restUser) DeleteAccount(w http.ResponseWriter, r *http.Request) error {
// 	userID := convert.IDUint(r)
// 	if userID == 0 {
// 		return handling.NewHttpError404(nil, "Pengguna tidak ditemukan.")
// 	}

// 	req := deleteUserForm{}
// 	if err := validator.DecodeValidate(r.Body, &req); err != nil {
// 		return err
// 	}
// 	e := req.toEntity()
// 	e.ID = userID
// 	if err := c.apiUser.DeleteAccount(r.Context(), e); err != nil {
// 		return err
// 	}
// 	response.SuccessData("user berhasil dihapus", "id", e.ID, w)
// 	return nil
// }

// func (c *restUser) RestoreAccount(w http.ResponseWriter, r *http.Request) error {
// 	userID := convert.IDUint(r)
// 	if userID == 0 {
// 		return handling.NewHttpError404(nil, "Pengguna tidak ditemukan.")
// 	}
// 	e := &entity.User{BaseModel: entity.BaseModel{ID: userID}}
// 	if err := c.apiUser.RestoreAccount(r.Context(), e); err != nil {
// 		return err
// 	}
// 	response.SuccessData("user berhasil di restore", "id", e.ID, w)
// 	return nil
// }

// func (c *restUser) DownloadFile(w http.ResponseWriter, r *http.Request) error {
// 	ctx := r.Context()

// 	userID := convert.IDUint(r)
// 	if userID == 0 {
// 		return handling.NewHttpError404(nil, "Pengguna tidak ditemukan.")
// 	}

// 	filename := r.PathValue("filename")
// 	if err := validator.IsValidFilename(filename); err != nil {
// 		return err
// 	}

// 	loggedInUserID, _ := middleware.GetUserIDFromContext(ctx)
// 	loggedInUserPerms, _ := middleware.GetBigPermsFromContext(ctx)
// 	hasRightAccess := loggedInUserPerms.HasOneOf(permissionindex.SuperAdmin, permissionindex.ReadUser)

// 	if !hasRightAccess && loggedInUserID != userID {
// 		return handling.NewHttpError(nil, http.StatusBadRequest, "Anda tidak memiliki akses untuk mengunduh berkas ini.", constant.ERR_VALIDATION_ERROR)
// 	}

// 	user := &entity.User{BaseModel: entity.BaseModel{ID: userID}}
// 	if err := c.apiUser.FindUserByID(r.Context(), user); err != nil {
// 		return err
// 	}

// 	imageFilenames := []string{
// 		user.FilePhotoKtp,
// 		user.FileScanSignature,
// 		user.FileVerificationKtp,
// 	}

// 	if user.ProfilePictureURL != nil {
// 		imageFilenames = append(imageFilenames, *user.ProfilePictureURL)
// 	}

// 	filenames := []string{
// 		user.FileFormNDA,
// 		user.FileFormUAA01,
// 		user.FilePaktaIntregritas,
// 		user.FileVideoLiveness,
// 	}
// 	filenames = append(filenames, imageFilenames...)

// 	if !slices.Contains(filenames, filename) {
// 		return handling.NewHttpError404(nil, "Berkas tidak ditemukan.")
// 	}

// 	buff, err := c.apiUser.GetFile(r.Context(), fmt.Sprintf("registration/%s", filename))
// 	if err != nil {
// 		return err
// 	}

// 	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
// 	w.Header().Set("Content-Type", "application/octet-stream")

// 	if slices.Contains(imageFilenames, filename) {
// 		watermarkedBuff, err := watermark.AddWatermark(buff, user.Fullname.String())
// 		if err == nil {
// 			buff = watermarkedBuff
// 		}
// 	}

// 	_, err = w.Write(buff)
// 	if err != nil {
// 		return handling.NewHttpError500(err)
// 	}

// 	return nil
// }

// func (c *restUser) FindUsersByIDs(w http.ResponseWriter, r *http.Request) error {
// 	req := findUsersByIDs{}

// 	if err := validator.DecodeValidate(r.Body, &req); err != nil {
// 		return err
// 	}

// 	users, err := c.apiUser.FindUsersByIDs(r.Context(), req.IDs)
// 	if err != nil {
// 		return err
// 	}

// 	response.SuccessData("users by ids sukses", "users", users, w)

// 	return nil
// }

// func (c *restUser) FindByUserID(w http.ResponseWriter, r *http.Request) error {
// 	userID := convert.IDUint(r)

// 	url := r.URL.Query()

// 	e := &entity.UserFindItemAPI{
// 		User: &entity.User{
// 			BaseModel: entity.BaseModel{ID: userID},
// 		},
// 	}

// 	includeKSO := true // default includeKSO to true
// 	if url.Has("include_kso") {
// 		includeKSO = slices.Contains([]string{"true", "1"}, strings.ToLower(url.Get("include_kso")))
// 	}
// 	if includeKSO {
// 		e.SetMemo("include_kso", includeKSO)
// 	}

// 	err := c.apiUser.FindUserByIDWithRoles(r.Context(), e)
// 	if err != nil {
// 		return err
// 	}

// 	response.SuccessData("user ditemukan", "user", e, w)
// 	return nil
// }

func (c *restUser) FindAll(w http.ResponseWriter, r *http.Request) error {
	var err error

	query := r.URL.Query()

	pagination := pagination.NewPagination(query)

	options := entity.UserFindAllOptions{
		IsDeleted: false,
		Status:    query.Get("status"),
		Name:      query.Get("name"),
		Phone:     query.Get("phone"),
		Email:     query.Get("email"),
		HasPerms:  query.Get("has_perms"),
		OrderBy:   query.Get("order_by"),
	}

	isDeletedStr := strings.TrimSpace(query.Get("is_deleted"))
	if isDeletedStr == "true" || isDeletedStr == "1" {
		options.IsDeleted = true
	}

	roleIDStr := strings.TrimSpace(query.Get("role_id"))
	if roleIDStr != "" {
		roleUUID, validRoleUUID := uuid.Parse(roleIDStr)
		if validRoleUUID != nil {
			return handling.NewHttpError(nil, http.StatusBadRequest, "role_id is invalid.", constant.ERR_VALIDATION_ERROR)
		}
		options.RoleID = roleUUID
	}

	err = c.apiUser.FindAll(r.Context(), pagination, options)
	if err != nil {
		return err
	}
	response.SuccessDataPaginate("success find all", "users", pagination.Rows, &response.Page{
		TotalData:   pagination.TotalRows,
		Limit:       pagination.GetLimit(),
		CurrentPage: pagination.GetPage(),
		TotalPage:   pagination.TotalPages,
	}, w)
	return nil
}

// func (c *restUser) ApproveRegistration(w http.ResponseWriter, r *http.Request) error {
// 	userID := convert.IDUint(r)
// 	if userID == 0 {
// 		return handling.NewHttpError404(nil, "Pengguna tidak ditemukan.")
// 	}

// 	var err error

// 	e := &entity.User{BaseModel: entity.BaseModel{ID: userID}}
// 	if err = c.apiUser.ApproveRegistration(r.Context(), e); err != nil {
// 		return err
// 	}

// 	digitalSignAccount := e.GetDigitalSignAccount("placeholder")
// 	if digitalSignAccount != nil {
// 		response.Success("User berhasil diverifikasi. Menunggu verifikasi digital sign.", w)
// 		return nil
// 	}

// 	response.Success("User berhasil diverifikasi.", w)
// 	return nil
// }

// func (c *restUser) RejectRegistration(w http.ResponseWriter, r *http.Request) error {
// 	userID := convert.IDUint(r)
// 	if userID == 0 {
// 		return handling.NewHttpError404(nil, "Pengguna tidak ditemukan.")
// 	}

// 	var err error

// 	req := updateRegistrationStatusForm{}
// 	if err = validator.DecodeValidate(r.Body, &req); err != nil {
// 		return err
// 	}

// 	e := req.toEntity()
// 	e.ID = userID
// 	e.RejectedReason = &req.Reason

// 	if err = c.apiUser.RejectRegistration(r.Context(), e); err != nil {
// 		return err
// 	}
// 	response.Success("pendaftaran user ditolak", w)
// 	return nil
// }

// func (c *restUser) GetSessions(w http.ResponseWriter, r *http.Request) error {
// 	ctx := r.Context()
// 	userId, err := middleware.GetUserIDFromContext(ctx)
// 	if err != nil {
// 		return err
// 	}

// 	authClaims, err := middleware.GetAuthClaimsFromContext(ctx)
// 	if err != nil {
// 		return err
// 	}

// 	sessions, err := c.apiUser.GetSessions(ctx, userId)
// 	if err != nil {
// 		return err
// 	}

// 	var results []map[string]any
// 	for _, e := range sessions {
// 		results = append(results, map[string]any{
// 			"id":               e.ID,
// 			"created_at":       e.CreatedAt,
// 			"updated_at":       e.UpdatedAt,
// 			"deleted_at":       e.DeletedAt,
// 			"session_id":       e.SessionID,
// 			"user_id":          e.UserID,
// 			"user_agent":       e.UserAgent,
// 			"ip_address":       e.IpAddress,
// 			"last_activity_at": e.LastActivityAt,
// 			"is_you":           authClaims.SessionID == e.SessionID,
// 		})
// 	}

// 	response.SuccessData("ok", "sessions", results, w)
// 	return nil
// }

// func (c *restUser) DeleteSession(w http.ResponseWriter, r *http.Request) error {
// 	ctx := r.Context()
// 	userId, err := middleware.GetUserIDFromContext(ctx)
// 	if err != nil {
// 		return err
// 	}
// 	authClaims, err := middleware.GetAuthClaimsFromContext(ctx)
// 	if err != nil {
// 		return err
// 	}

// 	sessionId := r.PathValue("id")

// 	if sessionId == authClaims.SessionID {
// 		return handling.NewHttpError(nil, http.StatusBadRequest, "session aktif, tidak dapat dihapus", constant.ERR_USER_SESSION_ACTIVE)
// 	}
// 	err = c.apiUser.DeleteSession(r.Context(), &entity.DeviceLog{UserID: userId, SessionID: sessionId})
// 	if err != nil {
// 		return err
// 	}
// 	response.Success("session berhasil di hapus", w)
// 	return nil
// }

// func (c *restUser) Suspend(w http.ResponseWriter, r *http.Request) error {
// 	userID := convert.IDUint(r)

// 	var err error

// 	req := updateRegistrationStatusForm{}
// 	if err = validator.DecodeValidate(r.Body, &req); err != nil {
// 		return err
// 	}

// 	e := req.toEntity()
// 	e.ID = userID
// 	e.SuspendedReason = &req.Reason

// 	if err = c.apiUser.Suspend(r.Context(), e); err != nil {
// 		return err
// 	}
// 	response.Success("user berhasil di suspend", w)
// 	return nil
// }

// func (c *restUser) ReActivate(w http.ResponseWriter, r *http.Request) error {
// 	userID := convert.IDUint(r)

// 	var err error
// 	e := &entity.User{BaseModel: entity.BaseModel{ID: userID}}

// 	if err = c.apiUser.ReActivate(r.Context(), e); err != nil {
// 		return err
// 	}
// 	response.Success("user account berhasil di aktifkan kembali", w)
// 	return nil
// }
