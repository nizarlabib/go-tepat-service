package middleware

import (
	"context"
	"net/http"

	"go-memolab-service/internal/entity"
	"go-memolab-service/internal/framework/secondary/repository"
	"go-memolab-service/internal/spesial"
	"go-memolab-service/internal/value_object/v1/handling"
	"go-memolab-service/internal/value_object/v1/router"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func RefreshTokenMiddleware(repo repository.User) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return router.RootHandler(func(w http.ResponseWriter, r *http.Request) error {
			ctx := r.Context()

			tokenString, err := spesial.ParseBearerToken(r)
			if err != nil {
				return err
			}

			claims := &spesial.Payload{}
			token, err := jwt.ParseWithClaims(tokenString, claims, spesial.GetJWTKeyFunc())

			if err != nil || !token.Valid {
				return handling.NewHttpError401(err)
			}

			if claims.SessionID == "" || claims.Type != spesial.PayloadTypeRefresh {
				return handling.NewHttpError401(err)
			}

			user := entity.NewUserFromID(claims.ID)
			if err = repo.FindByID(ctx, user); err != nil {
				return handling.NewHttpError401(err)
			}

			if user.RememberToken == nil || *user.RememberToken == "" {
				return handling.NewHttpError401(err)
			}

			if err = bcrypt.CompareHashAndPassword([]byte(*user.RememberToken), []byte(claims.SessionID)); err != nil {
				return handling.NewHttpError401(err)
			}

			if user.AccountStatus == entity.UserAccountStatusSuspended {
				return handling.NewHttpError401(nil)
			}

			ctx = spesial.AddAuthClaimsToContext(ctx, claims)
			ctx = spesial.AddUserIDToContext(ctx, claims.ID)
			ctx = addRefreshTokenToContext(ctx, tokenString)

			next.ServeHTTP(w, r.WithContext(ctx))
			return nil
		})
	}
}

const (
	authRefreshTokenKey spesial.ContextKey = "auth_refresh_token"
)

func addRefreshTokenToContext(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, authRefreshTokenKey, token)
}

func GetRefreshTokenFromContext(ctx context.Context) (string, error) {
	refreshToken, ok := ctx.Value(authRefreshTokenKey).(string)
	if !ok {
		return "", handling.NewHttpError401(nil)
	}
	return refreshToken, nil
}
