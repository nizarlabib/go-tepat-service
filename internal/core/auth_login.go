package core

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	// "time"

	"go-starter-service/internal/entity"
	"go-starter-service/internal/pkg/rbac/permission"
	"go-starter-service/internal/spesial"
	"go-starter-service/internal/value_object/v1/constant"
	"go-starter-service/internal/value_object/v1/convert"
	"go-starter-service/internal/value_object/v1/handling"

	permissionindex "go-starter-service/internal/pkg/rbac/permission-index"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidUserName = errors.New("username is required")
	ErrInvalidPassword = errors.New("password is required")
)

func (c *auth) processLogin(ctx context.Context, e *entity.Auth) error {
	if e.GetUsername() == "" {
		return ErrInvalidUserName
	}

	if e.Password == "" {
		return ErrInvalidPassword
	}

	reqInfo, err := spesial.GetRequestInfoFromContext(ctx)
	if err != nil {
		return err
	}
	rateLimitKey := fmt.Sprintf("rate:ip:%s", reqInfo.IPAddr)

	var user *entity.User

	user, err = c.attemptLoginWithEmailPass(ctx, e, rateLimitKey)
	if err != nil {
		return err
	}

	if err = c.errIfTooManyLoginAttempts(ctx, user, rateLimitKey); err != nil {
		return err
	}

	_ = c.redis.Del(ctx, rateLimitKey).Err()

	if err = c.validateUserRole(ctx, e, user); err != nil {
		return err
	}

	if err = c.saveUserDevice(ctx, user); err != nil {
		return err
	}

	if err = c.saveUserLoginActivity(ctx, user); err != nil {
		return err
	}

	err = c.generateAndInjectAuthToken(ctx, e, user)
	if err != nil {
		return err
	}

	return nil
}

func (c *auth) saveUserDevice(ctx context.Context, e *entity.User) error {
	reqInfo, err := spesial.GetRequestInfoFromContext(ctx)
	if err != nil {
		return err
	}

	sessionId := uuid.NewString()
	e.UserDevice = &entity.DeviceLog{
		SessionID:      sessionId,
		UserID:         e.ID,
		UserAgent:      reqInfo.UserAgent,
		IpAddress:      reqInfo.IPAddr,
		LastActivityAt: time.Now(),
	}

	if err := c.coreDeviceLogs.Create(ctx, e.UserDevice); err != nil {
		return err
	}

	return nil
}

func (c *auth) saveUserLoginActivity(ctx context.Context, e *entity.User) error {
	reqInfo, err := spesial.GetRequestInfoFromContext(ctx)
	if err != nil {
		return err
	}

	now := time.Now()

	rememberTokenBytes := make([]byte, 32)
	if _, err := rand.Read(rememberTokenBytes); err != nil {
		return handling.NewHttpError500(err)
	}
	rememberToken := hex.EncodeToString(rememberTokenBytes)

	hashedToken, err := bcrypt.GenerateFromPassword([]byte(rememberToken), 12)
	if err != nil {
		return handling.NewHttpError500(err)
	}
	e.RememberToken = convert.ToPointerStr(string(hashedToken))
	e.LastLoggedInAt = convert.ToPointerTime(now)
	e.LastLoggedInIP = &reqInfo.IPAddr
	e.LastLoggedInUserAgent = &reqInfo.UserAgent
	if e.Role != nil {
		e.LastLoggedInRoleID = &e.Role.ID
	}

	if err = c.repoUser.UpdateLoginActivity(ctx, e); err != nil {
		return err
	}

	e.UserDevice.RememberToken = rememberToken

	return nil
}

func (c *auth) errIfTooManyLoginAttempts(ctx context.Context, user *entity.User, rateLimitKey string) error {
	// Check Rate Limit
	attemptStr, err := c.redis.Get(ctx, rateLimitKey).Result()
	if err != nil {
		attemptStr = "0"
	}
	attempts, _ := strconv.Atoi(attemptStr)
	if attempts >= c.maxLoginAttempts {
		if user != nil {
			frozenReason := "too many login attempts"
			user.SuspendedReason = &frozenReason
			err = c.Suspend(ctx, user)
			if err != nil {
				return err
			}
		}
		_ = c.redis.Del(ctx, rateLimitKey).Err()
		errMsg := fmt.Sprintf("You have made %d failed login attempts. Your account has been suspended for a while. Please contact administrator.", attempts)
		return handling.NewHttpError(nil, http.StatusTooManyRequests, errMsg, constant.ERR_LOGIN_ATTEMPT_LIMIT)
	}

	return nil
}

func (c *auth) attemptLoginWithEmailPass(ctx context.Context, e *entity.Auth, rateLimitKey string) (*entity.User, error) {
	username := e.GetUsernameAsEmail()

	if username == "" {
		return nil, ErrInvalidUserName
	}
	if e.Password == "" {
		return nil, ErrInvalidPassword
	}

	var err error

	user := &entity.User{Email: username}
	if err = c.repoUser.FindByEmail(ctx, user); err != nil {
		return nil, err
	}

	if user.IsSuspended() {
		errMsg := "user account is suspended"
		if user.SuspendedReason != nil {
			errMsg = fmt.Sprintf("%s karena '%s'", errMsg, *user.SuspendedReason)
		}
		return nil, handling.NewHttpError(nil, http.StatusBadRequest, errMsg, constant.ERR_USER_SUSPENDED)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(e.Password)); err != nil {
		attemptStr, err := c.redis.Get(ctx, rateLimitKey).Result()
		if err != nil {
			attemptStr = "0"
		}
		attempts, _ := strconv.Atoi(attemptStr)

		if attempts < c.maxLoginAttempts {
			// Increase user's attempts to redis
			if err := c.redis.Incr(ctx, rateLimitKey).Err(); err != nil {
				return nil, handling.NewHttpError500(err)
			}
			attempts++
		}

		if attempts >= c.maxLoginAttempts {
			return user, nil
		}

		errMsg := fmt.Sprintf("Password is incorrect. You have %d attempts remaining.", c.maxLoginAttempts-attempts)
		return nil, handling.NewHttpError(nil, http.StatusUnauthorized, errMsg, constant.ERR_INVALID_CREDENTIALS)
	}

	return user, nil
}

func (c *auth) validateUserRole(ctx context.Context, _ *entity.Auth, user *entity.User) error {
	hasRoles := user.RoleID != nil
	if !hasRoles {
		return handling.NewHttpError(nil, http.StatusForbidden, "You don't have any valid role to login", constant.ERR_USER_ROLE_MISSED)
	}

	var role *entity.Role

	// check existing authenticated role from claims payload
	payload, err := spesial.GetAuthClaimsFromContext(ctx)
	if err == nil {
		if user.Role != nil {
			user.Role.PermissionMask = payload.PermissionMask
			role = user.Role
		}
	}

	if role == nil && user.LastLoggedInRoleID != nil {
		if user.Role != nil {
			role = user.Role
		}
	}

	if role == nil {
		role = user.Role
	}

	if !user.IsAccountVerified() {
		permissions := permission.New()

		if role.Can(permissionindex.SuperAdmin) {
			permissions.Grant(permissionindex.SuperAdmin)
		}
		if role.Can(permissionindex.Admin) {
			permissions.Grant(permissionindex.Admin)
		}
		if role.Can(permissionindex.User) {
			permissions.Grant(permissionindex.User)
		}

		role.SetPermissions(permissions)
	}

	user.Role = role

	return nil
}

func (c *auth) generateAuthTokenClaims(ctx context.Context, e *entity.User) *spesial.Payload {
	newClaims := &spesial.Payload{
		ID:             e.ID,
		RoleID:         e.Role.ID,
		Role:           e.Role.Name,
		PermissionMask: e.Role.GetPermissions().String(),
	}

	// carry over impersonation info if any
	authClaims, _ := spesial.GetAuthClaimsFromContext(ctx)

	var sessionID string

	if authClaims != nil && authClaims.SessionID != "" {
		sessionID = authClaims.SessionID
	} else if e.UserDevice != nil && e.UserDevice.RememberToken != "" {
		sessionID = e.UserDevice.RememberToken
	} else if e.UserDevice != nil && e.UserDevice.SessionID != "" {
		sessionID = e.UserDevice.SessionID
	} else {
		sessionID = uuid.NewString()
	}
	newClaims.SessionID = sessionID

	if authClaims != nil && authClaims.IsImpersonated() {
		newClaims.ImpersonatedBy = authClaims.ImpersonatedBy
		newClaims.ImpersonatedAt = authClaims.ImpersonatedAt
	}

	return newClaims
}

func (c *auth) generateAuthToken(_ context.Context, claims *spesial.Payload, e *entity.Auth) error {
	accessToken, err := spesial.GenerateAccessToken(claims)
	if err != nil {
		return err
	}
	e.AccessToken = accessToken

	refreshToken, err := spesial.GenerateRefreshToken(claims)
	if err != nil {
		return err
	}
	e.RefreshToken = refreshToken

	return nil
}

func (c *auth) generateAndInjectAuthToken(ctx context.Context, e *entity.Auth, user *entity.User) error {
	claims := c.generateAuthTokenClaims(ctx, user)

	if err := c.generateAuthToken(ctx, claims, e); err != nil {
		return err
	}

	return nil
}
