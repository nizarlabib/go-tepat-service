package pagination

import (
	"net/url"
	"strconv"
)

type Pagination struct {
	Limit int `json:"limit,omitempty"`

	Page       int `json:"page,omitempty"`
	TotalPages int `json:"total_pages"`

	Query map[string]string `json:"queries,omitempty"`

	TotalRows int64       `json:"total_rows"`
	Rows      interface{} `json:"rows"`
}

func NewPagination(query url.Values) *Pagination {
	limitStr := query.Get("limit")
	pageStr := query.Get("page")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}
	
	if limit > 50 {
		limit = 50
	}

	return &Pagination{
		Limit: limit,
		Page:  page,
	}
}

func (p *Pagination) GetOffset() int {
	return (p.GetPage() - 1) * p.GetLimit()
}
func (p *Pagination) GetPage() int {
	if p.Page <= 0 {
		p.Page = 1
	}
	return p.Page
}
func (p *Pagination) GetLimit() int {
	if p.Limit <= 0 {
		p.Limit = 10
	}
	if p.Limit > 100 {
		p.Limit = 100
	}
	return p.Limit
}

func (p *Pagination) QueryGet(key string) string {
	if p.Query == nil {
		return ""
	}
	return p.Query[key]
}

func (p *Pagination) QueryAdd(key, value string) {
	if p.Query == nil {
		p.Query = make(map[string]string)
	}
	p.Query[key] = value
}
