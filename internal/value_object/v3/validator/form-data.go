package validator

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"slices"

	"go-starter-service/internal/value_object/v1/constant"
	"go-starter-service/internal/value_object/v1/handling"
	"go-starter-service/internal/value_object/v1/response"
)

func DecodeFormData(w http.ResponseWriter, r *http.Request, fileNames []string) error {
	for _, field := range fileNames {
		uploadedFile, fileHeader, err := r.FormFile(field)
		if err != nil {
			return handling.NewHttpError(err, http.StatusBadRequest, fmt.Sprintf("Kolom %s wajib diisi dan harus berupa berkas.", field), constant.ERR_INVALID_FILE_FORMAT)
		}
		if !slices.Contains(constant.ValidFileUploadTypes, fileHeader.Header.Get("Content-Type")) {
			return handling.NewHttpError(nil, http.StatusBadRequest, fmt.Sprintf("Format berkas %s tidak didukung.", field), constant.ERR_INVALID_FILE_FORMAT)
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
