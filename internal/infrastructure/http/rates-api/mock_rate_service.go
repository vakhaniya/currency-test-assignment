package rates_api

import (
	"currency-rate-app/internal/domains/currency"
)

type MockRateService struct{}

func NewMockRateService() *MockRateService {
	return &MockRateService{}
}

var rates = map[string]float64{
	"USD": 0.5,
	"EUR": 1.0,
	"MXN": 1.5,
}

func (s *MockRateService) FetchData(baseCurrency currency.CurrencyCode) (map[string]float64, error) {
	return rates, nil
}
