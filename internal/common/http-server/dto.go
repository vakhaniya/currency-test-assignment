package http_server

import (
	error_utils "currency-rate-app/internal/common/error-utils"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
)

type HttpErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message,omitempty"`
}

func SendErrorResponse(w http.ResponseWriter, err error) {
	var response HttpErrorResponse
	var status int

	var customerErr *error_utils.CustomError
	if errors.As(err, &customerErr) {
		status = getHttpStatusByErrorType(customerErr.ErrorType)
		response = HttpErrorResponse{
			Code:    customerErr.Code,
			Message: customerErr.Message,
		}
	} else {
		status = http.StatusInternalServerError
		response = HttpErrorResponse{
			Code: "InternalServerError",
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("failed to encode error response", "error", err.Error())
	}
}

func SendSuccessResponse(w http.ResponseWriter, res any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func getHttpStatusByErrorType(errorType error_utils.ErrorType) int {
	switch errorType {
	case error_utils.ErrorCodeNotFound:
		return http.StatusNotFound
	case error_utils.ErrorCodeBadRequest:
		return http.StatusBadRequest
	case error_utils.ErrorCodeBusinessLogic:
		return http.StatusUnprocessableEntity
	case error_utils.ErrorCodeUnexpected:
		return http.StatusInternalServerError
	}

	return http.StatusInternalServerError
}
