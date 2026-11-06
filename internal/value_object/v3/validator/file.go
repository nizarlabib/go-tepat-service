package validator

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"slices"

	"go-starter-service/internal/entity"
	"go-starter-service/internal/value_object/v1/constant"
	"go-starter-service/internal/value_object/v1/convert"
	"go-starter-service/internal/value_object/v1/handling"

	"github.com/spf13/viper"
)

func NewFileValidator(form *multipart.Form) *fileValidator {
	return &fileValidator{
		files: make([]*entity.FileUpload, 0),
		form:  form,
	}
}

type fileValidator struct {
	files []*entity.FileUpload
	form  *multipart.Form
}

func (f *fileValidator) ValidateOptional(key string, expectedExts []string, maxFileSizeInMB int64) (filename string, err error) {
	return f.validate(true, key, expectedExts, maxFileSizeInMB)
}

func (f *fileValidator) Validate(key string, expectedExts []string, maxFileSizeInMB int64) (filename string, err error) {
	return f.validate(false, key, expectedExts, maxFileSizeInMB)
}

func (f *fileValidator) ValidateOtherFile(key string, expectedExts []string, maxFileSizeInMB int64, path string) (filename string, err error) {
	return f.validateOtherFile(true, key, expectedExts, maxFileSizeInMB, path)
}

func (f *fileValidator) validate(isOptional bool, key string, expectedExts []string, maxFileSizeInMB int64) (filename string, err error) {
	fhs := f.form.File[key]
	if len(fhs) == 0 {
		if isOptional {
			return "", nil
		}
		return "", handling.NewHttpError(nil, http.StatusBadRequest, fmt.Sprintf("Kolom %s wajib diisi dan harus berupa berkas.", key), constant.ERR_INVALID_FILE_FORMAT)
	}
	fh := fhs[0]

	contentType := fh.Header.Get("Content-Type")

	if !slices.Contains(constant.ValidFileUploadTypes, contentType) {
		return "", handling.NewHttpError(nil, http.StatusBadRequest, fmt.Sprintf("Format berkas %s tidak didukung.", key), constant.ERR_INVALID_FILE_FORMAT)
	}

	fileExt := constant.GetFileExtFromMime(contentType)
	if !slices.Contains(expectedExts, fileExt) {
		return "", handling.NewHttpError(nil, http.StatusBadRequest, fmt.Sprintf("Format berkas %s tidak didukung.", key), constant.ERR_INVALID_FILE_FORMAT)
	}

	fileSize := fh.Size
	if fileSize > (maxFileSizeInMB * (1 << 20)) {
		return "", handling.NewHttpError(nil, http.StatusBadRequest, fmt.Sprintf("Ukuran berkas %s terlalu besar.", key), constant.ERR_FILE_SIZE_TOO_BIG)
	}

	path := "registration"

	file := &entity.File{
		Name: convert.GenerateFileNameWithExt(fileExt),
		Mime: contentType,
		Size: fileSize,
		Path: path,
		Tags: []string{
			viper.GetString("APP_SERVICE_NAME"),
			path,
		},
	}
	fileUpload := &entity.FileUpload{
		FileInfo:    file,
		Path:        path,
		ObjectName:  fmt.Sprintf("%s/%s", file.Path, file.Name),
		MFileHeader: fh,
		Field:       key,
	}
	f.files = append(f.files, fileUpload)
	return file.Name, nil
}

func (f *fileValidator) validateOtherFile(isOptional bool, key string, expectedExts []string, maxFileSizeInMB int64, path string) (filename string, err error) {
	fhs := f.form.File[key]
	if len(fhs) == 0 {
		if isOptional {
			return "", nil
		}
		return "", handling.NewHttpError(nil, http.StatusBadRequest, fmt.Sprintf("Kolom %s wajib diisi dan harus berupa berkas.", key), constant.ERR_INVALID_FILE_FORMAT)
	}
	fh := fhs[0]

	contentType := fh.Header.Get("Content-Type")

	if !slices.Contains(constant.ValidFileUploadTypes, contentType) {
		return "", handling.NewHttpError(nil, http.StatusBadRequest, fmt.Sprintf("Format berkas %s tidak didukung.", key), constant.ERR_INVALID_FILE_FORMAT)
	}

	fileExt := constant.GetFileExtFromMime(contentType)
	if !slices.Contains(expectedExts, fileExt) {
		return "", handling.NewHttpError(nil, http.StatusBadRequest, fmt.Sprintf("Format berkas %s tidak didukung.", key), constant.ERR_INVALID_FILE_FORMAT)
	}

	fileSize := fh.Size
	if fileSize > (maxFileSizeInMB * (1 << 20)) {
		return "", handling.NewHttpError(nil, http.StatusBadRequest, fmt.Sprintf("Ukuran berkas %s terlalu besar.", key), constant.ERR_FILE_SIZE_TOO_BIG)
	}

	file := &entity.File{
		Name: convert.GenerateFileNameWithExt(fileExt),
		Mime: contentType,
		Size: fileSize,
		Path: path,
		Tags: []string{
			viper.GetString("APP_SERVICE_NAME"),
			path,
		},
	}
	fileUpload := &entity.FileUpload{
		FileInfo:    file,
		Path:        path,
		ObjectName:  fmt.Sprintf("%s/%s", file.Path, file.Name),
		MFileHeader: fh,
		Field:       key,
	}
	f.files = append(f.files, fileUpload)
	return file.Name, nil
}

func (f *fileValidator) Files() []*entity.FileUpload {
	return f.files
}
