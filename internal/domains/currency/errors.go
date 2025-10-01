package currency

import (
	error_utils "currency-rate-app/internal/common/error-utils"
)

func ErrInvalidCurrencyCode() *error_utils.CustomError {
	return &error_utils.CustomError{
		ErrorType: error_utils.ErrorCodeBusinessLogic,
		Code:      "InvalidCurrencyCode",
	}
}

func ErrCurrenciesShouldDiffer() *error_utils.CustomError {
	return &error_utils.CustomError{
		ErrorType: error_utils.ErrorCodeBusinessLogic,
		Code:      "CurrenciesShouldDiffer",
	}
}

func ErrCurrencyRateIdIsRequried() *error_utils.CustomError {
	return &error_utils.CustomError{
		ErrorType: error_utils.ErrorCodeBadRequest,
		Code:      "CurrencyRateIdIsRequried",
	}
}

func ErrCurrencyRateNotFound() *error_utils.CustomError {
	return &error_utils.CustomError{
		ErrorType: error_utils.ErrorCodeNotFound,
		Code:      "CurrencyRateNotFound",
	}
}

func ErrCurrencyRateNotCompletedYet() *error_utils.CustomError {
	return &error_utils.CustomError{
		ErrorType: error_utils.ErrorCodeBusinessLogic,
		Code:      "CurrencyRateNotCompletedYet",
	}
}

func ErrCurrencyRateFetchFailed() *error_utils.CustomError {
	return &error_utils.CustomError{
		ErrorType: error_utils.ErrorCodeBusinessLogic,
		Code:      "CurrencyRateFetchFailed",
	}
}
