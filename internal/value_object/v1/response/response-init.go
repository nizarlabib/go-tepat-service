package response

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type ResponseMeta struct {
	Message   string `json:"message"`
	Code      int    `json:"code"`
	ErrorCode string `json:"error_code,omitempty"`
	Detail    string `json:"detail,omitempty"`
}

// // TODO: DELETE THIS
type Response2 struct {
	Status  string                 `json:"status,omitempty"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data"`
	//Data    interface{} `json:"data,omitempty"`
}

type Response struct {
	Data interface{}  `json:"data,omitempty"`
	Meta ResponseMeta `json:"meta,omitempty"`
	//Data    interface{} `json:"data,omitempty"`
}

type PaginationResponse struct {
	//Message string                 `json:"message"`
	Data map[string]interface{} `json:"data"`
	Page Page                   `json:"page"`
	Meta ResponseMeta           `json:"meta,omitempty"`
}

type Page struct {
	CurrentPage int   `json:"current_page" json:"current_page,omitempty"`
	TotalData   int64 `json:"total_data" json:"total_data,omitempty"`
	Limit       int   `json:"limit" json:"limit,omitempty"`
	TotalPage   int   `json:"total_page" json:"total_page,omitempty"`
}

func (m *Page) SetPage(page, limit int, totalData int64) {
	m.CurrentPage = page
	m.Limit = limit
	m.TotalData = totalData

	if totalData == 0 {
		m.TotalPage = 1
		return
	}

	m.TotalPage = int(totalData) / limit
	if int(totalData)%limit > 0 {
		m.TotalPage++
	}

	if m.TotalPage == 0 {
		m.TotalPage = 1
	}

}

func JSON(response interface{}, statusCode int, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("error", slog.Any("error", err))
		return
	}
}
