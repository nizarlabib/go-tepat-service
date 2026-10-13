package validator

import (
	"encoding/json"
	"fmt"
	"go-memolab-service/internal/value_object/v1/constant"
	"go-memolab-service/internal/value_object/v1/handling"
	"go-memolab-service/internal/value_object/v1/response"
	"io"
	"mime/multipart"
	"net/http"
	"slices"
)

func DecodeFormData(w http.ResponseWriter, r *http.Request, fileNames []string) error {
	for _, field := range fileNames {
		uploadedFile, fileHeader, err := r.FormFile(field)
		if err != nil {
			return handling.NewHttpError(err, http.StatusBadRequest, fmt.Sprintf("field %s is required or it's not a file", field), constant.ERR_INVALID_FILE_FORMAT)
		}
		if !slices.Contains(constant.ValidFileUploadTypes, fileHeader.Header.Get("Content-Type")) {
			return handling.NewHttpError(nil, http.StatusBadRequest, fmt.Sprintf("file: %s format is not supported", field), constant.ERR_INVALID_FILE_FORMAT)
		}
		defer func(file multipart.File) {
			err := file.Close()
			if err != nil {
				response.Error(r.Context(), err, w)
			}
		}(uploadedFile)
	}

	return nil
}

func Decode(b io.Reader, v any) error {
	err := json.NewDecoder(b).Decode(v)
	if err != nil {
		return handling.NewHttpError(err, http.StatusBadRequest, "request body is invalid", constant.ERR_JSON_BODY)
	}
	return nil
}

func Request(s any) error {
	if err := validator(s); err != nil {
		return err
	}
	return nil
}
