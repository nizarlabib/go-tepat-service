package rbac

import (
	"fmt"
	"go-memolab-service/internal/entity"
	"go-memolab-service/internal/spesial"
	"go-memolab-service/internal/value_object/v1/handling"
	"go-memolab-service/internal/value_object/v1/response"
	"net/http"

	"github.com/google/uuid"
)

type permissionOption func(user *entity.User) bool

/**
* Generate UUID with 12 digits padded with leading zeros
 */
func UUIDPad(n int) uuid.UUID {
	if n < 1 || n > 281474976710655 { // 12 digits max
		panic("UUIDPad: number out of range, must be between 1 and 281474976710655")
	}
	uuidStr := fmt.Sprintf("00000000-0000-0000-0000-%012x", n)
	return uuid.MustParse(uuidStr)
}

func MustHasAll(bitIndices ...int) permissionOption {
	return func(user *entity.User) bool {
		return user.Can(bitIndices...)
	}
}

func MustHasOneOf(bitIndices ...int) permissionOption {
	return func(user *entity.User) bool {
		return user.MaybeCan(bitIndices...)
	}
}

func MustSuperAdmin() permissionOption {
	return func(user *entity.User) bool {
		return user.IsSuperAdmin()
	}
}

func WithPermission(opts ...permissionOption) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			user, err := spesial.GetUserFromContext(r.Context())
			if err != nil {
				response.Error(
					r.Context(),
					handling.NewHttpError401(err),
					w,
				)
				return
			}

			if len(opts) > 0 {
				for _, opt := range opts {
					if !opt(user) {
						// continue // need test
						response.Error(r.Context(), handling.NewHttpError403(err), w)
						return
					}
				}
			}

			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}
