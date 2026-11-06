package rest

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"

	"go-memolab-service/internal/api"
	"go-memolab-service/internal/entity"
	"go-memolab-service/internal/value_object/v1/constant"
	"go-memolab-service/internal/value_object/v1/handling"
	"go-memolab-service/internal/value_object/v1/response"
	"go-memolab-service/internal/value_object/v1/router"
	"go-memolab-service/internal/value_object/v1/validator"

	"github.com/go-chi/chi/v5"
)

type File interface {
	BulkUpload(w http.ResponseWriter, r *http.Request) error

	Start(ctx context.Context) chi.Router
}

func NewFile(api api.File) File {
	return &fileRest{
		api: api,
	}
}

type fileRest struct {
	api api.File
}

func (c *fileRest) Start(ctx context.Context) chi.Router {
	r := chi.NewRouter()

	r.Method(http.MethodPost, "/{path}/bulk-upload", router.RootHandler(c.BulkUpload))
	r.Method(http.MethodGet, "/{path}/{filename}", router.RootHandler(c.DownloadFile))
	r.Method(http.MethodDelete, "/{path}/{filename}", router.RootHandler(c.DeleteFile))

	return r
}

func (c *fileRest) BulkUpload(w http.ResponseWriter, r *http.Request) error {
	path := r.PathValue("path")
	if !slices.ContainsFunc(entity.ValidFilesMapping, func(f entity.FileMap) bool {
		return f.Path == path
	}) {
		return handling.NewHttpError(nil, http.StatusBadRequest, fmt.Sprintf("path: %s is invalid", path), constant.ERR_VALIDATION_ERROR)
	}

	fieldStr := r.FormValue("fields")
	fields := strings.Split(fieldStr, ",")

	// update file header
	for _, field := range fields {
		_, fileHeader, err := r.FormFile(field)
		if err != nil {
			return handling.NewHttpError(err, http.StatusBadRequest, fmt.Sprintf("field %s is required or it's not a file", field), constant.ERR_INVALID_FILE_FORMAT)
		}

		contentType := fileHeader.Header.Get("Content-Type")

		if isGenericContentType(contentType) {
			file, err := fileHeader.Open()
			if err != nil {
				return handling.NewHttpError500(err)
			}
			defer file.Close()

			buffer := make([]byte, 512)

			n, err := file.Read(buffer)
			if err != nil && err != io.EOF {
				return handling.NewHttpError500(err)
			}

			contentType = http.DetectContentType(buffer[:n])

			// update Content-Type di fileHeader
			fileHeader.Header.Set("Content-Type", contentType)
		}
	}

	if len(fields) == 0 {
		return handling.NewHttpError(nil, http.StatusBadRequest, "fields is required", constant.ERR_VALIDATION_ERROR)
	}

	if err := validator.DecodeFormData(w, r, fields); err != nil {
		return err
	}

	var files []*entity.FileUpload
	for _, field := range fields {
		if fhs := r.MultipartForm.File[field]; len(fhs) > 0 {
			files = append(files, &entity.FileUpload{
				Field:       field,
				Path:        path,
				MFileHeader: fhs[0],
			})
		}
	}

	if err := c.api.UploadFiles(r.Context(), files); err != nil {
		return err
	}

	result := map[string]interface{}{}
	for idx, field := range fields {
		file := files[idx]
		result[field] = map[string]interface{}{
			"filename": file.ObjectName,
			"filesize": file.Size,
		}
	}

	response.SuccessData("files uploaded succesfully", "uploaded_files", result, w)
	return nil
}

func isGenericContentType(contentType string) bool {
	return contentType == "" ||
		contentType == "application/octet-stream" ||
		contentType == "binary/octet-stream"
}

func (h fileRest) DownloadFile(w http.ResponseWriter, r *http.Request) error {
	filename := r.PathValue("filename")
	path := r.PathValue("path")

	e := &entity.File{Name: filename, Path: path}

	if err := validator.IsValidFilename(filename); err != nil {
		return err
	}

	buff, err := h.api.GetFileObject(r.Context(), e)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Header().Set("Content-Type", "application/octet-stream")

	_, err = w.Write(buff)
	if err != nil {
		return handling.NewHttpError500(err)
	}

	return nil
}

func (h fileRest) DeleteFile(w http.ResponseWriter, r *http.Request) error {
	filename := r.PathValue("filename")
	path := r.PathValue("path")

	e := &entity.File{Name: filename, Path: path}
	if err := h.api.Delete(r.Context(), e); err != nil {
		return err
	}
	response.SuccessData("success", "file", e.Name, w)
	return nil
}
