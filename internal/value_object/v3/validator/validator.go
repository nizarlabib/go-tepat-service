package validator

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"reflect"
	"strings"

	"go-starter-service/internal/value_object/v1/constant"
	"go-starter-service/internal/value_object/v1/handling"

	"github.com/go-playground/validator/v10"
)

func Validate(v any) error {
	if err := myValidator.Validate.Struct(v); err != nil {
		return formatValidationError(err)
	}
	return nil
}

func Decode(b io.ReadCloser, v any) error {
	if b == nil {
		return nil
	}

	// validate value v must pointer
	if reflect.TypeOf(v).Kind() != reflect.Ptr {
		slog.Error("validator: v must be pointer", "type", reflect.TypeOf(v))
		return handling.NewHttpError500(nil)
	}

	if v == nil {
		return handling.NewHttpError500(nil)
	}

	defer b.Close()

	err := json.NewDecoder(b).Decode(v)
	if err != nil {
		return handling.NewHttpError(err, http.StatusBadRequest, "Request body is not valid JSON.", constant.ERR_JSON_BODY)
	}
	return nil
}

func DecodeValidate(b io.ReadCloser, v any) error {
	if err := Decode(b, v); err != nil {
		return err
	}

	if err := Validate(v); err != nil {
		return err
	}

	return nil
}

func formatValidationError(err error) error {
	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		slog.Error("validator: unable to cast validation error", "error", err)
		return handling.NewHttpError(err, http.StatusBadRequest, "Terjadi kesalahan validasi.", constant.ERR_VALIDATION_ERROR)
	}

	// immediately return the first error message
	fe := validationErrors[0]
	feMsg := fe.Translate(myValidator.Trans)

	ns := strings.Split(fe.Namespace(), ".")[1:] // remove top struct name
	fn := strings.Join(ns, ".")

	feMsg2 := strings.Replace(feMsg, fe.Field(), fn, 1)

	return handling.NewHttpError(nil, http.StatusBadRequest, feMsg2, constant.ERR_VALIDATION_ERROR)
}
