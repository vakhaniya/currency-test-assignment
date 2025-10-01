package rates_api

import (
	"currency-rate-app/internal/common/config"
	"currency-rate-app/internal/domains/currency"
	"net/http"
)

type RateService interface {
	FetchData(baseCurrency currency.CurrencyCode) (map[string]float64, error)
}

type RateServiceType string

const (
	Frankfurter RateServiceType = "Frankfurter"
	Mock        RateServiceType = "Mock"
)

func NewRateService(httpClient *http.Client, config config.Config) RateService {
	switch config.RatesApiType {
	case "Frankfurter":
		return NewFrankfurterRateService(httpClient, config.FrankfurterApiURL)
	case "Mock":
		return NewMockRateService()
	default:
		return NewMockRateService()
	}
}
