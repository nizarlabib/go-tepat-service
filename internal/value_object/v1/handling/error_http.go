package handling

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"go-starter-service/internal/value_object/v1/constant"

	"github.com/getsentry/sentry-go"
)

type ClientError interface {
	Error() string
	GetErrorCode() string
	GetMessage() string
	GetStatus() int
	GetFlagError() *FlagError
	GetErrorData() any
}

func IsClientError(err error) bool {
	_, ok := err.(ClientError)
	return ok
}

func NewHttpError500(err error) error {
	return NewHttpError(err, http.StatusInternalServerError, "Internal server error", constant.ERR_INTERNAL_SERVER_ERROR)
}

func NewHttpError400(err error, msg, errCode string) error {
	return NewHttpError(err, http.StatusBadRequest, msg, errCode)
}

func NewHttpError401(err error) error {
	return NewHttpError(err, http.StatusUnauthorized, "User unauthorized", constant.ERR_UNAUTHORIZED)
}

func NewHttpError403(err error) error {
	return NewHttpError(err, http.StatusForbidden, "Forbidden", constant.ERR_FORBIDDEN)
}

func NewHttpError404(err error, message string) error {
	return NewHttpError(err, http.StatusNotFound, message, constant.ERR_RECORD_NOT_FOUND)
}

func NewHttpError(err error, status int, message string, errorCode string) error {
	if err != nil && status == http.StatusInternalServerError {
		fmt.Printf("error occured: %v\n", err)
		debug.PrintStack()
		sentry.CaptureException(err)
	}

	return &HTTPError{
		Cause:     err,
		Status:    status,
		ErrorCode: errorCode,
		Message:   message,
	}
}

func NewHttpErrorWithFlag(err error, status int, message string, errorCode string, flagError *FlagError) error {
	return &HTTPError{
		Cause:     err,
		Status:    status,
		ErrorCode: errorCode,
		Message:   message,
		FlagError: flagError,
	}
}

func NewHttpErrorWithData(err error, status int, message string, errorCode string, data any) error {
	return &HTTPError{
		Cause:     err,
		Status:    status,
		ErrorCode: errorCode,
		Message:   message,
		Data:      data,
	}
}

type HTTPError struct {
	Cause     error
	Message   string
	ErrorCode string
	Status    int
	Data      any
	FlagError *FlagError
}

func (e *HTTPError) Error() string {
	if e.Cause == nil {
		return e.Message
	}
	return e.Message + " : " + e.Cause.Error()
}

func (e *HTTPError) GetMessage() string {
	if e.Message == "" {
		return e.Error()
	}
	return e.Message
}

func (e *HTTPError) GetStatus() int {
	return e.Status
}

func (e *HTTPError) GetErrorCode() string {
	return e.ErrorCode
}

type FlagError struct {
	Message string `json:"message,omitempty"`

	UserAgent    string    `json:"user_agent"`
	IpAddress    string    `json:"ip_address"`
	LastActivity time.Time `json:"last_activity"`
}

func (e *HTTPError) GetFlagError() *FlagError {
	return e.FlagError
}

func (e *HTTPError) GetErrorData() any {
	return e.Data
}
