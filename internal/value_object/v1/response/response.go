package response

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"

	"go-starter-service/internal/value_object/v1/constant"
	"go-starter-service/internal/value_object/v1/handling"
	"go-starter-service/internal/value_object/v1/utils/types"

	"github.com/getsentry/sentry-go"
)

var reportedCodes = types.NewSet(
	http.StatusInternalServerError,
	http.StatusBadGateway,
	http.StatusRequestTimeout,
	http.StatusServiceUnavailable,
	http.StatusGatewayTimeout,
	http.StatusTooManyRequests,
)

func Error(ctx context.Context, err error, w http.ResponseWriter) {
	if err == nil {
		return
	}

	var res Response
	clientErr, isClientErr := err.(handling.ClientError)

	timeoutRes := Response{
		Meta: ResponseMeta{
			Message:   "Waktu permintaan habis. Silakan coba lagi.",
			Code:      http.StatusRequestTimeout,
			ErrorCode: constant.ERR_REQUEST_TIMEOUT,
		},
	}

	if errors.Is(err, context.DeadlineExceeded) {
		res = timeoutRes
	} else if os.IsTimeout(err) {
		res = timeoutRes
	} else if isClientErr {
		if clientErr.GetErrorData() != nil {
			res = Response{
				Data: map[string]any{
					"error": clientErr.GetErrorData(),
				},
				Meta: ResponseMeta{
					Message:   clientErr.GetMessage(),
					Code:      clientErr.GetStatus(),
					ErrorCode: clientErr.GetErrorCode(),
				},
			}
		} else if clientErr.GetFlagError() != nil {
			res = Response{
				Data: map[string]any{
					"activity": clientErr.GetFlagError(),
				},
				Meta: ResponseMeta{
					Message:   clientErr.GetMessage(),
					Code:      clientErr.GetStatus(),
					ErrorCode: clientErr.GetErrorCode(),
				},
			}
		} else {
			res = Response{
				Meta: ResponseMeta{
					Message:   clientErr.GetMessage(),
					Code:      clientErr.GetStatus(),
					ErrorCode: clientErr.GetErrorCode(),
				},
			}
		}
	} else {
		slog.Error("http error occurred", slog.String("error", err.Error()))
		res = Response{
			Meta: ResponseMeta{
				Message:   "Internal server error.",
				Code:      http.StatusInternalServerError,
				ErrorCode: constant.ERR_INTERNAL_SERVER_ERROR,
			},
		}
	}

	if reportedCodes.Contains(res.Meta.Code) {
		slog.Error("reporting error to sentry", slog.String("error", err.Error()))
		sentryHub := sentry.GetHubFromContext(ctx)
		if sentryHub != nil {
			sentryHub.CaptureException(err)
		}
	}

	JSON(res, res.Meta.Code, w)
}

func Success(msg string, w http.ResponseWriter) {
	response := Response{
		Data: nil,
		Meta: ResponseMeta{
			Message: msg,
			Code:    http.StatusOK,
		},
	}
	JSON(response, http.StatusOK, w)
}

func SuccessCreate(msg, dataName string, data any, w http.ResponseWriter) {
	response := Response{
		Data: map[string]any{
			dataName: data,
		},
		Meta: ResponseMeta{
			Message: msg,
			Code:    http.StatusCreated,
		},
	}
	JSON(response, http.StatusCreated, w)
}

func SuccessDataPaginate(msg, dataName string, data any, page *Page, w http.ResponseWriter) {
	response := PaginationResponse{
		Data: map[string]any{
			dataName: data,
		},
		Meta: ResponseMeta{
			Message: msg,
			Code:    http.StatusOK,
		},
		Page: Page{
			Limit:       page.Limit,
			TotalPage:   page.TotalPage,
			TotalData:   page.TotalData,
			CurrentPage: page.CurrentPage,
		},
	}
	JSON(response, http.StatusOK, w)
}

func SuccessData(msg, dataName string, data any, w http.ResponseWriter) {
	response := Response{
		Data: map[string]any{
			dataName: data,
		},
		Meta: ResponseMeta{
			Message: msg,
			Code:    http.StatusOK,
		},
	}

	JSON(response, http.StatusOK, w)
}

func SuccessLogin(msg, accessToken, refreshToken string, w http.ResponseWriter) {
	response := Response{
		Data: map[string]any{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
		},
		Meta: ResponseMeta{
			Message: msg,
			Code:    http.StatusOK,
		},
	}

	JSON(response, http.StatusOK, w)
}
