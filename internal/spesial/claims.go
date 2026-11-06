package spesial

import (
	"context"

	"go-starter-service/internal/entity"
	"go-starter-service/internal/pkg/rbac/permission"
	"go-starter-service/internal/value_object/v1/handling"
)

type ContextKey string

const (
	authUserIDKey ContextKey = "auth_user_id"
	authUserKey   ContextKey = "auth_user"
	authClaims    ContextKey = "auth_claims"
	authBigPerms  ContextKey = "auth_big_perms"
)

func AddAuthClaimsToContext(ctx context.Context, claims *Payload) context.Context {
	ctx = context.WithValue(ctx, authClaims, claims)

	ctx = AddUserIDToContext(ctx, claims.ID)

	var bigPerms *permission.BigPerm
	if claims.PermissionMask != "" {
		bigPerms = permission.NewFromStr(claims.PermissionMask)
	}

	if bigPerms != nil {
		ctx = AddBigPermsToContext(ctx, bigPerms)
	}

	return ctx
}

func GetAuthClaimsFromContext(ctx context.Context) (*Payload, error) {
	res, ok := ctx.Value(authClaims).(*Payload)
	if !ok && res == nil {
		return nil, handling.NewHttpError401(nil)
	}
	return res, nil
}

func AddUserToContext(ctx context.Context, e *entity.User) context.Context {
	return context.WithValue(ctx, authUserKey, e)
}

func GetUserFromContext(ctx context.Context) (*entity.User, error) {
	user, ok := ctx.Value(authUserKey).(*entity.User)
	if !ok && user == nil {
		return nil, handling.NewHttpError401(nil)
	}
	return user, nil
}

func AddBigPermsToContext(ctx context.Context, perms *permission.BigPerm) context.Context {
	return context.WithValue(ctx, authBigPerms, perms)
}

func GetBigPermsFromContext(ctx context.Context) (*permission.BigPerm, error) {
	perms, ok := ctx.Value(authBigPerms).(*permission.BigPerm)
	if !ok && perms == nil {
		return nil, handling.NewHttpError401(nil)
	}
	return perms, nil
}

func AddUserIDToContext(ctx context.Context, userID uint) context.Context {
	return context.WithValue(ctx, authUserIDKey, userID)
}

func GetUserIDFromContext(ctx context.Context) (uint, error) {
	userID, ok := ctx.Value(authUserIDKey).(uint)
	if !ok && userID == 0 {
		return 0, handling.NewHttpError401(nil)
	}
	return userID, nil
}
