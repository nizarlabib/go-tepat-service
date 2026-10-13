package ratelimiter

import (
	"net/http"
	"time"

	"go-memolab-service/internal/value_object/v1/constant"
	"go-memolab-service/internal/value_object/v1/response"

	"github.com/go-chi/httprate"
)

func Middleware(rate int, period time.Duration) func(http.Handler) http.Handler {
	limitHandler := func(w http.ResponseWriter, r *http.Request) {
		res := response.Response{
			Meta: response.ResponseMeta{
				Message:   "Terlalu banyak permintaan.",
				Code:      http.StatusTooManyRequests,
				ErrorCode: constant.ERR_TOO_MANY_REQUESTS,
			},
		}

		response.JSON(res, http.StatusTooManyRequests, w)
	}

	keyFuncs := []httprate.KeyFunc{
		httprate.KeyByIP,
		httprate.KeyByEndpoint,
		httprate.KeyByRealIP,
	}

	return httprate.Limit(
		rate,
		period,
		httprate.WithKeyFuncs(keyFuncs...),
		httprate.WithLimitHandler(limitHandler),
	)
}
