package repository

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go-memolab-service/internal/entity"
	"go-memolab-service/internal/framework/secondary/storage"
	"go-memolab-service/internal/infrastructure"
	"go-memolab-service/internal/value_object/v1/constant"
	"go-memolab-service/internal/value_object/v1/handling"
	"go-memolab-service/internal/value_object/v1/pagination"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User interface {
	Delete(ctx context.Context, e *entity.User) error
	Restore(ctx context.Context, e *entity.User) error
	Custom(ctx context.Context, fn func(db *gorm.DB) *gorm.DB) error
	CountNeedAction(ctx context.Context) (int64, error)
	FindAll(ctx context.Context, pg *pagination.Pagination, opts entity.UserFindAllOptions) error
	FindByPhone(ctx context.Context, e *entity.User) error
	FindByEmail(ctx context.Context, e *entity.User) error
	FindByIDWithRoles(ctx context.Context, e *entity.UserFindItemAPI) error
	FindByID(ctx context.Context, e *entity.User) error
	FindByIDs(ctx context.Context, ids []uint) ([]*entity.UserFindItemAPI, error)
	VerifyExistEmail(ctx context.Context, e *entity.User) error
	VerifyExistPhone(ctx context.Context, e *entity.User) error
	Register(ctx context.Context, e *entity.User) error
	UpdateUser(ctx context.Context, e *entity.User, updateColumns []string) error
	Save(ctx context.Context, e *entity.User) error
	ReplaceRole(ctx context.Context, e *entity.User, roles ...*entity.Role) error
	UpdateUserStatus(ctx context.Context, e *entity.User) error
	UpdateLoginActivity(ctx context.Context, e *entity.User) error
	RevokeAuthentication(ctx context.Context, e *entity.User) error
	UpdatePassword(ctx context.Context, e *entity.User) error
	GetSessions(ctx context.Context, userId uint) ([]*entity.DeviceLog, error)
	GetSessionByID(ctx context.Context, e *entity.DeviceLog) error
	UpdateSession(ctx context.Context, e *entity.DeviceLog) error
	DeleteSession(ctx context.Context, e *entity.DeviceLog) error
	BlockToken(ctx context.Context, jti string, ttl int) error
	IsTokenBlocked(ctx context.Context, jti string) (bool, error)
}

func NewUser(infra infrastructure.Postgres, redisClient *redis.Client, storage storage.Storage) User {
	db := infra.GetClientGorm()
	return &user{
		db:            db,
		client:        redisClient,
		fileRepo:      &file{db: db},
		storage:       storage,
		deviceLogRepo: &deviceLog{db: db},
	}
}

type user struct {
	db            *gorm.DB
	fileRepo      *file
	client        *redis.Client
	storage       storage.Storage
	deviceLogRepo *deviceLog
}

func (r *user) Delete(ctx context.Context, e *entity.User) error {
	shouldChangeEmail := e.Email != "" && !strings.HasSuffix(e.Email, "deleted")
	if shouldChangeEmail {
		e.Email = fmt.Sprintf("%s.deleted", e.Email)
	}

	result := r.db.WithContext(ctx).
		Select([]string{
			"email",
			"account_deletion_reason",
			"deleted_at",
			"deleted_by",
			"remember_token",
		}).
		Updates(e)

	if result.Error != nil {
		return handling.NewHttpError500(result.Error)
	}

	if result.RowsAffected == 0 {
		return handling.NewHttpError404(nil, "Pengguna tidak ditemukan.")
	}

	return nil
}

func (r *user) Restore(ctx context.Context, e *entity.User) error {
	result := r.db.WithContext(ctx).
		Model(e).
		Where("id = ? AND deleted_at IS NOT NULL", e.ID).
		Update("deleted_at", nil)

	if result.Error != nil {
		return handling.NewHttpError500(result.Error)
	}

	if result.RowsAffected == 0 {
		return handling.NewHttpError404(nil, fmt.Sprintf("Pengguna dengan ID %d tidak ditemukan atau belum dihapus.", e.ID))
	}

	return nil
}

func (r *user) Custom(ctx context.Context, fn func(db *gorm.DB) *gorm.DB) error {
	if err := fn(r.db.WithContext(ctx)).Error; err != nil {
		return err
	}
	return nil
}

func (r *user) CountNeedAction(ctx context.Context) (int64, error) {
	pg := pagination.NewPagination(make(url.Values))

	opts := entity.UserFindAllOptions{
		Status:    string(entity.UserAccountStatusPending),
		CountOnly: true,
	}

	err := r.FindAll(ctx, pg, opts)
	if err != nil {
		return 0, err
	}

	return pg.TotalRows, nil
}

func (r *user) FindAll(ctx context.Context, pg *pagination.Pagination, opts entity.UserFindAllOptions) error {
	var users []*entity.User

	var args []interface{}
	var conditions []string
	var values []interface{}

	if opts.Name != "" {
		conditions = append(conditions, "fullname ILIKE ?")
		values = append(values, "%"+opts.Name+"%")
	}

	if opts.Email != "" {
		conditions = append(conditions, "email ILIKE ?")
		values = append(values, "%"+opts.Email+"%")
	}

	if len(conditions) > 0 {
		args = append(args, strings.Join(conditions, " AND "))
		args = append(args, values...)
	}

	db := r.db.WithContext(ctx).Scopes(pagination.Paginate(users, pg, r.db, args...))

	if len(args) > 1 {
		db = db.Where(args[0], args[1:]...)
	}

	if err := db.Order("id ASC").Preload("Role").Find(&users).Error; err != nil {
		return handling.NewHttpError500(err)
	}

	userResults := make([]*entity.User, 0, len(users))
	for _, u := range users {
		user := u
		user.Permissions = u.Role.GetPermissions().String()
		userResults = append(userResults, user)
	}

	pg.Rows = users

	return nil
}

// FindByPhone implements [repository.User].
func (r *user) FindByPhone(ctx context.Context, e *entity.User) error {
	result := r.db.Model(e).
		Where("phone = ?", e.Phone).
		Preload("Role").
		First(&e)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return handling.NewHttpError404(
				result.Error,
				"Phone number is not registered.",
			)
		}
		return handling.NewHttpError500(result.Error)
	}

	return nil
}

func (r *user) FindByEmail(ctx context.Context, e *entity.User) error {
	result := r.db.Model(e).
		Where("users.email = ?", e.Email).
		Preload("Role").
		First(&e)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return handling.NewHttpError404(
				result.Error,
				"Email is not registered.",
			)
		}
		return handling.NewHttpError500(result.Error)
	}

	return nil
}

func (r *user) FindByIDWithRoles(ctx context.Context, e *entity.UserFindItemAPI) error {
	model := new(entity.User)
	model.ID = e.ID

	result := r.db.Model(model).
		Preload("Role").
		Where("users.id = ?", model.ID).
		First(model)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return handling.NewHttpError404(result.Error, fmt.Sprintf("User with ID %d not found.", e.ID))
		}
		return handling.NewHttpError500(result.Error)
	}

	return nil
}

func (r *user) FindByID(ctx context.Context, e *entity.User) error {
	result := r.db.Model(e).
		Preload("Role").
		Where("id = ?", e.ID).First(&e)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return handling.NewHttpError404(result.Error, fmt.Sprintf("User with ID %d not found.", e.ID))
		}
		return handling.NewHttpError500(result.Error)
	}

	if e.Role != nil {
		e.Permissions = e.Role.GetPermissions().String()
	}

	return nil
}

func (r *user) FindByIDs(ctx context.Context, ids []uint) ([]*entity.UserFindItemAPI, error) {
	var users []*entity.UserFindItemAPI

	e := &entity.User{}

	query := r.db.
		// Clauses(clause.Select{
		// 	Expression: e.SelectAllExpr("users", `
		// 		json_agg(
		// 			json_build_object(
		// 				'id', r.uuid,
		// 				'name', r.name,
		// 				'institute_id', i.id,
		// 				'institute_name', i.name,
		// 				'permission_mask', r.permission_mask
		// 			)
		// 		) AS roles,
		// 		json_build_object(
		// 			'id', i.id,
		// 			'parent_id', i.parent_id,
		// 			'name', i.name,
		// 			'is_contractor', i.is_contractor
		// 		) as institute
		// `),
		// }).
		Model(e).
		Joins("LEFT JOIN user_roles_v2 ur ON ur.user_id = users.id").
		Joins("LEFT JOIN roles r ON r.uuid = ur.role_id AND r.deleted_at IS NULL").
		Joins("LEFT JOIN institutes i ON i.id = users.institute_id").
		Where("users.id IN ?", ids).
		Group("users.id, i.id").
		Find(&users)

	err := query.Error
	if err != nil {
		return nil, handling.NewHttpError500(err)
	}

	return users, nil
}

func (r *user) VerifyExistEmail(ctx context.Context, e *entity.User) error {
	var count int64
	result := r.db.WithContext(ctx).
		Model(&entity.User{}).
		Unscoped().
		Where("email = ?", e.Email).
		Count(&count)

	if result.Error != nil {
		return handling.NewHttpError500(result.Error)
	}

	if count > 0 {
		return handling.NewHttpError(result.Error, http.StatusConflict, "Email already registered.", constant.ERR_EMAIL_EXISTS)
	}

	return nil
}

func (r *user) VerifyExistPhone(ctx context.Context, e *entity.User) error {
	var count int64
	result := r.db.WithContext(ctx).Model(&entity.User{}).
		Where("phone = ?", e.Phone).
		Count(&count)

	if result.Error != nil {
		return handling.NewHttpError500(result.Error)
	}

	if count > 0 {
		return handling.NewHttpError(result.Error, http.StatusConflict, "Phone number already registered.", constant.ERR_PHONE_EXISTS)
	}

	return nil
}

func (r *user) Register(ctx context.Context, e *entity.User) error {
	tx := r.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()
	defer tx.Rollback()

	if e.Password != "" && !strings.HasPrefix(e.Password, "$2a$") {
		hashedPass, _ := bcrypt.GenerateFromPassword([]byte(e.Password), 12)
		e.Password = string(hashedPass)
	}
	err := tx.Create(e).Error
	if err != nil {
		return err
	}

	tx.Commit()
	return nil
}

func (r *user) UpdateUser(ctx context.Context, e *entity.User, updateColumns []string) error {
	if len(updateColumns) == 0 {
		return nil
	}

	err := r.db.
		WithContext(ctx).
		Model(&entity.User{}).
		Select(updateColumns).
		Where("id = ?", e.ID).
		Updates(e).
		Error
	return err
}

func (r *user) Update(ctx context.Context, e *entity.User) error {
	err := r.db.WithContext(ctx).Updates(e).Error
	return err
}

func (r *user) Save(ctx context.Context, e *entity.User) error {
	err := r.db.WithContext(ctx).Save(e).Error
	return err
}

func (r *user) ReplaceRole(ctx context.Context, e *entity.User, roles ...*entity.Role) error {
	if len(roles) == 0 {
		return nil
	}

	err := r.db.WithContext(ctx).
		Model(e).
		Association("Roles").
		Replace(roles)
	return err
}

func (r *user) UpdateUserStatus(ctx context.Context, e *entity.User) error {
	err := r.db.WithContext(ctx).
		Select([]string{
			"rejected_at", "rejected_reason",
			"verified_at",
			"suspended_at", "suspended_reason",
			"rejected_at", "rejected_reason",
			"remember_token", "account_status",
		}).
		Updates(e).Error
	if err != nil {
		return handling.NewHttpError500(err)
	}

	return nil
}

func (r *user) UpdateLoginActivity(ctx context.Context, e *entity.User) error {
	columns := []string{
		"remember_token",
		"last_logged_in_at",
		"last_logged_in_ip",
		"last_logged_in_user_agent",
	}
	if e.LastLoggedInRoleID != nil && *e.LastLoggedInRoleID != uuid.Nil {
		columns = append(columns, "last_logged_in_role_id")
	}

	err := r.db.WithContext(ctx).
		Select(columns).
		Updates(e).Error
	if err != nil {
		return handling.NewHttpError500(err)
	}

	return nil
}

func (r *user) RevokeAuthentication(ctx context.Context, e *entity.User) error {
	e.RememberToken = nil

	err := r.db.WithContext(ctx).
		Select([]string{
			"remember_token",
		}).
		Updates(e).Error
	if err != nil {
		return handling.NewHttpError500(err)
	}

	return nil
}

func (r *user) UpdatePassword(ctx context.Context, e *entity.User) error {
	err := r.db.WithContext(ctx).
		Select("password", "password_changed_at", "password_must_change", "remember_token").
		Updates(e).
		Error
	if err != nil {
		return handling.NewHttpError500(err)
	}

	return nil
}

func (r *user) GetSessions(ctx context.Context, userId uint) ([]*entity.DeviceLog, error) {
	sessions, err := r.deviceLogRepo.FindAllByUserID(ctx, userId)
	if err != nil {
		return nil, err
	}
	return sessions, nil
}

func (c *user) GetSessionByID(ctx context.Context, e *entity.DeviceLog) error {
	if err := c.deviceLogRepo.FindByID(ctx, e); err != nil {
		return err
	}
	return nil
}

func (c *user) UpdateSession(ctx context.Context, e *entity.DeviceLog) error {
	if err := c.deviceLogRepo.Update(ctx, e); err != nil {
		return err
	}
	return nil
}

func (c *user) DeleteSession(ctx context.Context, e *entity.DeviceLog) error {
	if err := c.deviceLogRepo.Delete(ctx, e); err != nil {
		return err
	}
	return nil
}

func (r *user) BlockToken(ctx context.Context, jti string, ttl int) error {
	if jti == "" {
		return nil
	}
	key := fmt.Sprintf("blocklist:%s", jti)
	return r.client.Set(ctx, key, "1", time.Duration(ttl)*time.Second).Err()
}

func (r *user) IsTokenBlocked(ctx context.Context, jti string) (bool, error) {
	if jti == "" {
		return false, nil
	}
	key := fmt.Sprintf("blocklist:%s", jti)
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return val == "1", nil
}
