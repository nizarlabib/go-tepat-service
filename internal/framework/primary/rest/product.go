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

type Product interface {
	Start() chi.Router
}

func NewProduct(api api.Product) Product {
	return &product{api: api}
}

type product struct {
	api api.Product
}

func (c *product) Start() chi.Router {
	r := chi.NewRouter()

	r.Method(http.MethodGet, "/", router.RootHandler(c.findAll))
	r.Method(http.MethodGet, "/slug/{slug}", router.RootHandler(c.findBySlug))
	r.Method(http.MethodGet, "/{id}", router.RootHandler(c.findByID))
	r.Method(http.MethodPost, "/", router.RootHandler(c.create))
	r.Method(http.MethodPut, "/{id}", router.RootHandler(c.update))
	r.Method(http.MethodPatch, "/{id}/status", router.RootHandler(c.updateStatus))
	r.Method(http.MethodDelete, "/{id}", router.RootHandler(c.delete))

	return r
}

func (c *product) findAll(w http.ResponseWriter, r *http.Request) error {
	query := r.URL.Query()
	pg := pagination.NewPagination(query)

	payload := &entity.Product{
		Name:         query.Get("search"),
		CategorySlug: query.Get("category"),
	}

	if err := c.api.FindAll(r.Context(), payload, pg); err != nil {
		return err
	}
	response.SuccessDataPaginate("success", "products", pg.Rows, &response.Page{
		TotalData:   pg.TotalRows,
		Limit:       pg.GetLimit(),
		CurrentPage: pg.GetPage(),
		TotalPage:   pg.TotalPages,
	}, w)
	return nil
}

func (c *product) findBySlug(w http.ResponseWriter, r *http.Request) error {
	slug := r.PathValue("slug")
	if slug == "" {
		return handling.NewHttpError404(nil, "Product not found.")
	}
	product, err := c.api.FindBySlug(r.Context(), slug)
	if err != nil {
		return err
	}
	response.SuccessData("success", "product", product, w)
	return nil
}

func (c *product) findByID(w http.ResponseWriter, r *http.Request) error {
	id, err := convert.PathValueUUIDChi(r, "id")
	if err != nil {
		return err
	}
	product, err := c.api.FindByID(r.Context(), id)
	if err != nil {
		return err
	}
	response.SuccessData("success", "product", product, w)
	return nil
}

func (c *product) create(w http.ResponseWriter, r *http.Request) error {
	payload := &entity.Product{}
	if err := validator.Decode(r.Body, payload); err != nil {
		return err
	}
	if err := validator.Request(*payload); err != nil {
		return err
	}
	if err := c.api.Create(r.Context(), payload); err != nil {
		return err
	}
	response.SuccessCreate("success", "product", payload, w)
	return nil
}

func (c *product) update(w http.ResponseWriter, r *http.Request) error {
	id, err := convert.PathValueUUIDChi(r, "id")
	if err != nil {
		return err
	}
	payload := &entity.Product{
		BaseUUIDModel: entity.BaseUUIDModel{ID: id},
	}
	if err := validator.Decode(r.Body, payload); err != nil {
		return err
	}
	if err := validator.Request(*payload); err != nil {
		return err
	}
	if err := c.api.Update(r.Context(), id, payload); err != nil {
		return err
	}
	response.SuccessData("success", "product", payload, w)
	return nil
}

func (c *product) updateStatus(w http.ResponseWriter, r *http.Request) error {
	id, err := convert.PathValueUUIDChi(r, "id")
	if err != nil {
		return err
	}
	payload := &entity.Product{}
	if err := validator.Decode(r.Body, payload); err != nil {
		return err
	}
	if err := validator.Request(*payload); err != nil {
		return err
	}
	if err := c.api.UpdateStatus(r.Context(), id, payload.IsActive); err != nil {
		return err
	}
	response.Success("success", w)
	return nil
}

func (c *product) delete(w http.ResponseWriter, r *http.Request) error {
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
