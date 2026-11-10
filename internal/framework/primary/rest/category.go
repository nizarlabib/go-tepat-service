package rest

import (
	"go-starter-service/internal/api"
	"go-starter-service/internal/entity"
	"go-starter-service/internal/value_object/v1/convert"
	"go-starter-service/internal/value_object/v1/handling"
	"go-starter-service/internal/value_object/v1/pagination"
	"go-starter-service/internal/value_object/v1/response"
	"go-starter-service/internal/value_object/v1/router"
	"go-starter-service/internal/value_object/v1/validator"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Category interface {
	Start() chi.Router
}

func NewCategory(api api.Category) Category {
	return &category{api: api}
}

type category struct {
	api api.Category
}

func (c *category) Start() chi.Router {
	r := chi.NewRouter()

	r.Method(http.MethodGet, "/", router.RootHandler(c.findAll))
	r.Method(http.MethodGet, "/slug/{slug}", router.RootHandler(c.findBySlug))
	r.Method(http.MethodGet, "/{id}", router.RootHandler(c.findByID))
	r.Method(http.MethodPost, "/", router.RootHandler(c.create))
	r.Method(http.MethodPut, "/{id}", router.RootHandler(c.update))
	r.Method(http.MethodDelete, "/{id}", router.RootHandler(c.delete))

	return r
}

func (c *category) findAll(w http.ResponseWriter, r *http.Request) error {
	query := r.URL.Query()
	pg := pagination.NewPagination(query)

	payload := &entity.Category{
		Name: query.Get("search"),
	}

	if err := c.api.FindAll(r.Context(), payload, pg); err != nil {
		return err
	}
	response.SuccessDataPaginate("success", "categories", pg.Rows, &response.Page{
		TotalData:   pg.TotalRows,
		Limit:       pg.GetLimit(),
		CurrentPage: pg.GetPage(),
		TotalPage:   pg.TotalPages,
	}, w)
	return nil
}

func (c *category) findBySlug(w http.ResponseWriter, r *http.Request) error {
	slug := r.PathValue("slug")
	if slug == "" {
		return handling.NewHttpError404(nil, "Category not found.")
	}
	category, err := c.api.FindBySlug(r.Context(), slug)
	if err != nil {
		return err
	}
	response.SuccessData("success", "category", category, w)
	return nil
}

func (c *category) findByID(w http.ResponseWriter, r *http.Request) error {
	id, err := convert.PathValueUUIDChi(r, "id")
	if err != nil {
		return err
	}
	category, err := c.api.FindByID(r.Context(), id)
	if err != nil {
		return err
	}
	response.SuccessData("success", "category", category, w)
	return nil
}

func (c *category) create(w http.ResponseWriter, r *http.Request) error {
	payload := &entity.Category{}
	if err := validator.Decode(r.Body, payload); err != nil {
		return err
	}
	if err := validator.Request(*payload); err != nil {
		return err
	}
	if err := c.api.Create(r.Context(), payload); err != nil {
		return err
	}
	response.SuccessCreate("success", "category", payload, w)
	return nil
}

func (c *category) update(w http.ResponseWriter, r *http.Request) error {
	id, err := convert.PathValueUUIDChi(r, "id")
	if err != nil {
		return err
	}
	payload := &entity.Category{
		BaseUUIDModel: entity.BaseUUIDModel{ID: id},
	}
	if err := validator.Decode(r.Body, payload); err != nil {
		return err
	}
	if err := validator.Request(*payload); err != nil {
		return err
	}
	if err := c.api.Update(r.Context(), payload); err != nil {
		return err
	}
	response.SuccessData("success", "category", payload, w)
	return nil
}

func (c *category) delete(w http.ResponseWriter, r *http.Request) error {
	id, err := convert.PathValueUUIDChi(r, "id")
	if err != nil {
		return err
	}
	if err := c.api.Delete(r.Context(), id); err != nil {
		return err
	}
	response.Success("success", w)
	return nil
}
