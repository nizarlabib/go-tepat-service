package middleware

import (
	"context"
	"net/http"

	// "slices"
	// "time"

	"go-memolab-service/internal/entity"
	"go-memolab-service/internal/spesial"
	"go-memolab-service/internal/value_object/v1/handling"
	"go-memolab-service/internal/value_object/v1/router"

	"github.com/dgrijalva/jwt-go"
)

func RefreshTokenMiddleware(ctx context.Context) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return router.RootHandler(func(w http.ResponseWriter, r *http.Request) error {
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
			// if err := core.FindUserByID(ctx, user); err != nil {
			// 	return handling.NewHttpError401(err)
			// }

			if user.RememberToken == nil {
				return handling.NewHttpError401(err)
			}

			switch user.AccountStatus {
			case entity.UserAccountStatusSuspended:
				return handling.NewHttpError401(nil)
			}

			// sessions, err := core.GetSessions(ctx, user.ID)
			// if err != nil {
			// 	return err
			// }

			// sessionIdx := slices.IndexFunc(sessions, func(s *entity.DeviceLog) bool {
			// 	return s.SessionID == claims.SessionID
			// })

			// // if sessionIdx < 0 {
			// // 	return handling.NewHttpError401(nil)
			// // }

			// if sessionIdx > -1 {
			// 	session := sessions[sessionIdx]
			// 	session.LastActivityAt = time.Now()
			// 	api.UpdateSession(ctx, session)
			// }

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
