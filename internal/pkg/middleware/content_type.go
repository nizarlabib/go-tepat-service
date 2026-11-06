package middleware

import (
	"net/http"
	"strings"
)

func ContentType(allowed ...string) func(http.Handler) http.Handler {
	allowedMethods := map[string]bool{
		http.MethodPost:  true,
		http.MethodPut:   true,
		http.MethodPatch: true,
	}

	allowedSet := make(map[string]bool, len(allowed))
	for _, a := range allowed {
		allowedSet[a] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if allowedMethods[r.Method] {
				contentType := r.Header.Get("Content-Type")
				if contentType == "" {
					http.Error(w, "Content-Type header is required", http.StatusUnsupportedMediaType)
					return
				}
				mediaType := strings.Split(contentType, ";")[0]
				mediaType = strings.TrimSpace(mediaType)
				if !allowedSet[mediaType] {
					http.Error(w, "Unsupported Content-Type", http.StatusUnsupportedMediaType)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
