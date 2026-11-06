package router

import (
	"net/http"

	"go-starter-service/internal/value_object/v1/response"
)

type RootHandler func(http.ResponseWriter, *http.Request) error

func (h RootHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h(w, r)
	if err == nil {
		return
	}

	response.Error(r.Context(), err, w)
}
