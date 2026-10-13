package spesial

import (
	"context"
	"net/http"
	"time"

	"go-memolab-service/internal/value_object/v1/constant"
	"go-memolab-service/internal/value_object/v1/handling"
)

const (
	requestInfoKey ContextKey = "app_req_info"
)

type RequestInfo struct {
	IPAddr    string    `json:"ip_addr"`
	UserAgent string    `json:"user_agent"`
	Location  string    `json:"location"`
	Endpoint  string    `json:"endpoint"`
	Timestamp time.Time `json:"timestamp"`
}

func getReqInfoFromRequest(r *http.Request) RequestInfo {
	return RequestInfo{
		IPAddr:    r.RemoteAddr,
		UserAgent: r.UserAgent(),
		Location:  "unknown",
		Endpoint:  r.URL.Path,
		Timestamp: time.Now(),
	}
}

func WithRequestInfo(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := AddRequestInfoToContext(r.Context(), getReqInfoFromRequest(r))
		h.ServeHTTP(w, r.WithContext(ctx))
	}

	return http.HandlerFunc(fn)
}

func AddRequestInfoToContext(ctx context.Context, requestInfo RequestInfo) context.Context {
	return context.WithValue(ctx, requestInfoKey, requestInfo)
}

func GetRequestInfoFromContext(ctx context.Context) (*RequestInfo, error) {
	requestInfo, ok := ctx.Value(requestInfoKey).(RequestInfo)
	if !ok {
		return nil, handling.NewHttpError(nil, http.StatusBadRequest, "Permintaan tidak valid.", constant.ERR_FORBIDDEN)
	}
	return &requestInfo, nil
}
