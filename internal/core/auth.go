package core

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"go-memolab-service/internal/entity"
	"go-memolab-service/internal/framework/secondary/repository"
	"go-memolab-service/internal/value_object/v1/constant"

	// "go-memolab-service/internal/value_object/v1/convert"
	"go-memolab-service/internal/value_object/v1/handling"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
)

type Auth interface {
	UploadFiles(ctx context.Context, files []*entity.FileUpload) error
	DeleteFiles(ctx context.Context, files []*entity.FileUpload) error
	FindUserByID(ctx context.Context, e *entity.User) error
	CheckExistingUserByEmail(ctx context.Context, e *entity.User) error
	CheckExistingUserByPhone(ctx context.Context, e *entity.User) error
	Register(ctx context.Context, e *entity.User, files []*entity.FileUpload, roles []*entity.Role) error
	Login(ctx context.Context, e *entity.Auth) error
	RefreshToken(ctx context.Context, e *entity.Auth) error
	Logout(ctx context.Context, e *entity.User) error

	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token, password string) error

	GenerateCsrfToken(ctx context.Context, e *entity.Auth) error
}

const (
	UserType     = "user"
	ProviderType = "provider"
)

type authOptions func(fac *auth)

func NewAuth(opts ...authOptions) Auth {
	viper.SetDefault("AUTH_LOGIN_MAX_ATTEMPTS", 5)

	authMaxAttempts := viper.GetInt("AUTH_LOGIN_MAX_ATTEMPTS")

	fac := &auth{
		maxLoginAttempts: authMaxAttempts,
	}
	for _, opt := range opts {
		opt(fac)
	}

	return fac
}

func AuthWithRepo(repo repository.User) authOptions {
	return func(fac *auth) {
		fac.repoUser = repo
	}
}

func AuthWithRedis(redis *redis.Client) authOptions {
	return func(fac *auth) {
		fac.redis = redis
	}
}

func AuthWithDeviceLog(deviceLog DeviceLog) authOptions {
	return func(fac *auth) {
		fac.coreDeviceLogs = deviceLog
	}
}

type auth struct {
	repoUser         repository.User
	redis            *redis.Client
	maxLoginAttempts int
	coreDeviceLogs   DeviceLog
}

// func (c *auth) GetActivityLogs(ctx context.Context, pg *pagination.Pagination) error {
// 	user, err := middleware.GetUserFromContext(ctx)
// 	if err != nil {
// 		return err
// 	}
// 	return c.coreActivityLog.FindAll(ctx, user, pg)
// }

func (c *auth) UploadFiles(ctx context.Context, files []*entity.FileUpload) error {
	// return c.repoUser.UploadFiles(ctx, files)
	return nil
}

func (c *auth) DeleteFiles(ctx context.Context, files []*entity.FileUpload) error {
	// return c.repoUser.DeleteFiles(ctx, files)
	return nil
}

func (c *auth) FindUserByID(ctx context.Context, e *entity.User) error {
	return c.repoUser.FindByID(ctx, e)
}

func (c *auth) CheckExistingUserByEmail(ctx context.Context, e *entity.User) error {
	return c.repoUser.VerifyExistEmail(ctx, e)
}

func (c *auth) CheckExistingUserByPhone(ctx context.Context, e *entity.User) error {
	return c.repoUser.VerifyExistPhone(ctx, e)
}

func (c *auth) Register(ctx context.Context, e *entity.User, files []*entity.FileUpload, roles []*entity.Role) error {
	if err := c.repoUser.Register(ctx, e); err != nil {
		return err
	}

	return nil
}

func (c *auth) Login(ctx context.Context, e *entity.Auth) error {
	err := c.processLogin(ctx, e)
	if err != nil {
		if handling.IsClientError(err) {
			return err
		} else {
			acceptedErrors := []error{
				ErrInvalidUserName,
				ErrInvalidPassword,
			}
			for _, acceptedErr := range acceptedErrors {
				if errors.Is(err, acceptedErr) {
					return handling.NewHttpError(err, http.StatusUnprocessableEntity, err.Error(), constant.ERR_VALIDATION_ERROR)
				}
			}
		}
		return handling.NewHttpError500(err)
	}
	return nil
}

func (c *auth) RefreshToken(ctx context.Context, e *entity.Auth) error {
	user := new(entity.User)
	user.ID = e.UserID

	// e2 := &entity.UserFindItemAPI{
	// 	User: user,
	// }

	// if err := c.repoUser.FindByIDWithRoles(ctx, e2); err != nil {
	// 	return err
	// }

	if user.IsSuspended() {
		return handling.NewHttpError(nil, http.StatusBadRequest, fmt.Sprintf("User account is suspended because: '%s'", *user.SuspendedReason), constant.ERR_USER_SUSPENDED)
	}

	if err := c.validateUserRole(ctx, e, user); err != nil {
		return err
	}

	// if user.UserDevice == nil {
	// 	user.UserDevice = &entity.DeviceLog{SessionID: e.SessionID}
	// }
	err := c.generateAndInjectAuthToken(ctx, e, user)
	if err != nil {
		return err
	}

	return nil
}

func (c *auth) Logout(ctx context.Context, e *entity.User) error {
	if err := c.repoUser.FindByID(ctx, e); err != nil {
		return err
	}

	disableSingleSession := viper.GetBool("DISABLE_SINGLE_SESSION")
	if !disableSingleSession {
		if err := c.repoUser.RevokeAuthentication(ctx, e); err != nil {
			return err
		}
	}

	// spesial.LogUserActivity(ctx, c.activityLog, "user.logout", nil)
	return nil
}

func (c *auth) Suspend(ctx context.Context, e *entity.User) error {
	existingUser := new(entity.User)
	existingUser.ID = e.ID
	if err := c.repoUser.FindByID(ctx, existingUser); err != nil {
		return err
	}

	if existingUser.AccountStatus == entity.UserAccountStatusSuspended {
		return handling.NewHttpError(nil, http.StatusBadRequest, "User is already suspended", constant.ERR_USER_SUSPENDED)
	}

	now := time.Now()
	existingUser.SuspendedReason = e.SuspendedReason
	existingUser.SuspendedAt = &now
	existingUser.RememberToken = nil
	existingUser.AccountStatus = entity.UserAccountStatusSuspended
	err := c.repoUser.UpdateUserStatus(ctx, existingUser)
	if err != nil {
		return err
	}
	// cacheKey := fmt.Sprintf("user_me:%s", convert.UintToStr(existingUser.ID))
	// cacheKey2 := fmt.Sprintf("me:%s", convert.UintToStr(existingUser.ID))
	// cacheKey3 := fmt.Sprintf("user:%s", convert.UintToStr(existingUser.ID))
	// _ = cache.RevalidateKeys(ctx, c.redis, cacheKey, cacheKey2, cacheKey3)

	// changedFields := []string{"account_status", "suspended_at", "suspended_reason"}
	// _ = c.event.EmitWithAdditionalData(ctx, string(events.EventUserInfoUpdated), *existingUser, changedFields)

	// middleware.LogUserActivity(ctx, c.activityLog, "user.suspend", map[string]any{"user_id": existingUser.ID})

	return nil
}

func (c *auth) GenerateCsrfToken(ctx context.Context, e *entity.Auth) error {
	token, err := generateCSRFToken()
	if err != nil {
		return err
	}

	decToken, err := decryptToken(token)
	if err != nil {
		return err
	}

	csrfKey := fmt.Sprintf("csrf:%s", decToken)
	err = c.redis.Set(ctx, csrfKey, "valid", 30*time.Minute).Err()
	if err != nil {
		return handling.NewHttpError500(err)
	}

	e.CsrfToken = token

	return nil
}

func generateCSRFToken() (string, error) {
	aesGCM, err := aesGcm()
	if err != nil {
		return "", handling.NewHttpError500(err)
	}

	token := make([]byte, aes.BlockSize)
	_, err = io.ReadFull(rand.Reader, token)
	if err != nil {
		return "", handling.NewHttpError500(err)
	}

	nonce := make([]byte, aesGCM.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return "", handling.NewHttpError500(err)
	}

	encryptedToken := aesGCM.Seal(nonce, nonce, token, nil)
	return hex.EncodeToString(encryptedToken), nil
}

func decryptToken(encryptedToken string) (string, error) {
	aesGCM, err := aesGcm()
	if err != nil {
		return "", err
	}

	encryptedData, err := hex.DecodeString(encryptedToken)
	if err != nil {
		return "", err
	}

	nonceSize := aesGCM.NonceSize()
	if len(encryptedData) < nonceSize {
		return "", fmt.Errorf("invalid encrypted token")
	}
	nonce, ciphertext := encryptedData[:nonceSize], encryptedData[nonceSize:]

	decryptedToken, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(decryptedToken), nil
}

func aesGcm() (cipher.AEAD, error) {
	keyBase64 := viper.GetString("AES_KEY")
	if keyBase64 == "" {
		return nil, handling.NewHttpError500(errors.New("aes key not found"))
	}

	// Decode the hex-encoded key
	key, err := base64.URLEncoding.DecodeString(keyBase64)
	if err != nil {
		return nil, handling.NewHttpError500(err)
	}

	// Create a new AES cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	return cipher.NewGCM(block)
}

func (c *auth) ForgotPassword(ctx context.Context, email string) error {
	user := &entity.User{Email: email}
	if err := c.repoUser.FindByEmail(ctx, user); err != nil {
		return err
	}

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return handling.NewHttpError500(err)
	}
	token := hex.EncodeToString(tokenBytes)

	hashedToken, err := bcrypt.GenerateFromPassword([]byte(token), 12)
	if err != nil {
		return handling.NewHttpError500(err)
	}

	ttl := viper.GetInt("PASSWORD_RESET_TTL")
	if ttl <= 0 {
		ttl = 900
	}

	resetKey := fmt.Sprintf("password_reset:%s", email)
	if err := c.redis.Set(ctx, resetKey, string(hashedToken), time.Duration(ttl)*time.Second).Err(); err != nil {
		return handling.NewHttpError500(err)
	}

	reverseKey := fmt.Sprintf("password_reset_token:%s", token)
	if err := c.redis.Set(ctx, reverseKey, email, time.Duration(ttl)*time.Second).Err(); err != nil {
		return handling.NewHttpError500(err)
	}

	return nil
}

func (c *auth) ResetPassword(ctx context.Context, token, password string) error {
	reverseKey := fmt.Sprintf("password_reset_token:%s", token)
	email, err := c.redis.Get(ctx, reverseKey).Result()
	if err == redis.Nil {
		return handling.NewHttpError(nil, http.StatusBadRequest, "Invalid or expired reset token", constant.ERR_TOKEN_EXPIRED)
	}
	if err != nil {
		return handling.NewHttpError500(err)
	}

	resetKey := fmt.Sprintf("password_reset:%s", email)
	storedHash, err := c.redis.Get(ctx, resetKey).Result()
	if err == redis.Nil {
		return handling.NewHttpError(nil, http.StatusBadRequest, "Invalid or expired reset token", constant.ERR_TOKEN_EXPIRED)
	}
	if err != nil {
		return handling.NewHttpError500(err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(token)); err != nil {
		return handling.NewHttpError(nil, http.StatusBadRequest, "Invalid or expired reset token", constant.ERR_TOKEN_EXPIRED)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return handling.NewHttpError500(err)
	}

	user := &entity.User{Email: email}
	if err := c.repoUser.FindByEmail(ctx, user); err != nil {
		return err
	}

	user.Password = string(hashedPassword)
	if err := c.repoUser.UpdatePassword(ctx, user); err != nil {
		return err
	}

	if err := c.redis.Del(ctx, resetKey, reverseKey).Err(); err != nil {
		return handling.NewHttpError500(err)
	}

	return nil
}
