package currency

import (
	"currency-rate-app/internal/domains/currency"
	"time"
)

type CreateRateRequest struct {
	BaseCurrency   currency.CurrencyCode `json:"baseCurrency" validate:"required,len=3"`
	ResultCurrency currency.CurrencyCode `json:"resultCurrency" validate:"required,len=3"`
}

type CreateRateResponse struct {
	Id string `json:"id"`
}

type GetCurrencyResponse struct {
	BaseCurrency   currency.CurrencyCode `json:"baseCurrency"`
	ResultCurrency currency.CurrencyCode `json:"resultCurrency"`
	Rate           float64               `json:"rate"`
	CompletedAt    time.Time             `json:"completedAt"`
}

func ToGetCurrencyResponse(currencyRate currency.CurrencyRate) *GetCurrencyResponse {
	return &GetCurrencyResponse{
		BaseCurrency:   currencyRate.BaseCurrency,
		ResultCurrency: currencyRate.ResultCurrency,
		Rate:           *currencyRate.Rate,
		CompletedAt:    currencyRate.CompletedAt.UTC(),
	}
}
