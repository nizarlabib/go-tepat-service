package spesial

import (
	"net/http"
	"time"

	"go-memolab-service/internal/value_object/v1/convert"

	"github.com/getsentry/sentry-go"
)

func SentryUserMiddleware(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		hub := sentry.GetHubFromContext(ctx)
		if hub == nil {
			hub = sentry.CurrentHub().Clone()
			ctx = sentry.SetHubOnContext(ctx, hub)
		}

		var breadcrumb *sentry.Breadcrumb

		proxy, _ := GetRequestInfoFromContext(ctx)
		user, err := GetUserFromContext(ctx)
		if err == nil {
			hub.ConfigureScope(func(scope *sentry.Scope) {
				sentryUser := sentry.User{
					ID:       convert.UintToStr(user.ID),
					Email:    user.Email,
					Username: user.Email,
					Name:     user.Fullname,
					Data: map[string]string{
						"account_status": string(user.AccountStatus),
						"is_superadmin":  convert.BoolToStr(user.HasSuperAdminRole()),
					},
				}
				if proxy != nil {
					sentryUser.IPAddress = proxy.IPAddr
				}

				scope.SetUser(sentryUser)
				scope.SetTag("user_id", sentryUser.ID)
			})

			breadcrumb = &sentry.Breadcrumb{
				Category:  "http",
				Type:      "default",
				Message:   "User Authenticated",
				Timestamp: time.Now(),
			}
		} else {
			breadcrumb = &sentry.Breadcrumb{
				Category:  "http",
				Type:      "default",
				Message:   "User Not Authenticated",
				Timestamp: time.Now(),
			}
		}

		hub.AddBreadcrumb(breadcrumb, nil)

		h.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}
