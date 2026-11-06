package middleware

import (
	"context"
	"errors"
	"log"
	"net/http"
	"slices"
	"time"

	"go-memolab-service/internal/entity"
	"go-memolab-service/internal/framework/secondary/repository"
	"go-memolab-service/internal/spesial"
	"go-memolab-service/internal/value_object/v1/constant"
	"go-memolab-service/internal/value_object/v1/handling"
	"go-memolab-service/internal/value_object/v1/router"

	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
)

func AuthMiddleware(repo repository.User) func(http.Handler) http.Handler {
	disableSingleSession := viper.GetBool("DISABLE_SINGLE_SESSION")
	validateUserSession := func(ctx context.Context, user *entity.User, claims *spesial.Payload) error {
		if disableSingleSession {
			return nil
		}

		if user.RememberToken == nil || *user.RememberToken == "" {
			return handling.NewHttpError401(errors.New("unauthorized"))
		}

		if err := bcrypt.CompareHashAndPassword([]byte(*user.RememberToken), []byte(claims.SessionID)); err != nil {
			flagError := &handling.FlagError{
				IpAddress:    *user.LastLoggedInIP,
				UserAgent:    *user.LastLoggedInUserAgent,
				LastActivity: *user.LastLoggedInAt,
			}

			return handling.NewHttpErrorWithFlag(err, http.StatusUnauthorized, "User session has been used in another location", constant.ERR_SESSION_CONFLICT, flagError)
		}

		sessions, err := repo.GetSessions(ctx, user.ID)
		if err != nil {
			return err
		}

		sessionIdx := slices.IndexFunc(sessions, func(s *entity.DeviceLog) bool {
			return s.SessionID == claims.SessionID
		})

		if sessionIdx < 0 {
			return handling.NewHttpError401(nil)
		}

		if sessionIdx > -1 {
			session := sessions[sessionIdx]
			session.LastActivityAt = time.Now()
			_ = repo.UpdateSession(ctx, session)
		}
		return nil
	}

	validateUserToken := func(ctx context.Context, claims *spesial.Payload) (*entity.User, error) {
		var err error

		user := entity.NewUserFromID(claims.ID)
		if err = repo.FindByID(ctx, user); err != nil {
			return nil, handling.NewHttpError401(err)
		}

		claims.UserID = user.ID

		if claims.IsImpersonated() {
			return user, nil
		}

		if user.RememberToken == nil {
			return nil, handling.NewHttpError401(nil)
		}

		if user.AccountStatus == entity.UserAccountStatusSuspended {
			return nil, handling.NewHttpError401(nil)
		}

		if err = validateUserSession(ctx, user, claims); err != nil {
			return nil, err
		}

		return user, nil
	}

	authHandler := func(next http.Handler) http.Handler {
		return router.RootHandler(func(w http.ResponseWriter, r *http.Request) error {

			ctx := r.Context()

			tokenString, err := spesial.ParseBearerToken(r)
			if err != nil {
				return err
			}

			claims, err := spesial.ValidateToken(tokenString)
			if err != nil {
				return err
			}

			if blocked, err := repo.IsTokenBlocked(ctx, claims.JTI); err != nil {
				return handling.NewHttpError500(err)
			} else if blocked {
				return handling.NewHttpError401(errors.New("token has been revoked"))
			}

			ctx = spesial.AddAuthClaimsToContext(ctx, claims)

			var user *entity.User

			user, err = validateUserToken(ctx, claims)
			if err != nil {
				log.Println(err)
				return err
			}

			if user == nil {
				return handling.NewHttpError401(nil)
			}

			ctx = spesial.AddAuthClaimsToContext(ctx, claims)
			ctx = spesial.AddUserToContext(ctx, user)

			next.ServeHTTP(w, r.WithContext(ctx))
			return nil
		})
	}
	return func(h http.Handler) http.Handler {
		handler := spesial.SentryUserMiddleware(h)
		handler = authHandler(handler)
		return handler
	}
}
