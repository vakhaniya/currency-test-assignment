package currency

import (
	"time"
)

type CurrencyCode string

const (
	USD CurrencyCode = "USD"
	EUR CurrencyCode = "EUR"
	MXN CurrencyCode = "MXN"
)

func (s CurrencyCode) IsValid() bool {
	switch s {
	case USD, EUR, MXN:
		return true
	}

	return false
}

type CurrencyRateStatus string

const (
	CurrencyRateStatusPending    CurrencyRateStatus = "PENDING"
	CurrencyRateStatusProcessing CurrencyRateStatus = "PROCESSING"
	CurrencyRateStatusCompleted  CurrencyRateStatus = "COMPLETED"
	CurrencyRateStatusFailed     CurrencyRateStatus = "FAILED"
)

type CurrencyRate struct {
	Id             string
	IdempotencyKey string
	BaseCurrency   CurrencyCode
	ResultCurrency CurrencyCode
	Status         CurrencyRateStatus
	Rate           *float64
	CompletedAt    *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func ValidateCurrencyPair(baseCurrency CurrencyCode, resultCurrency CurrencyCode) error {
	if !baseCurrency.IsValid() {
		return ErrInvalidCurrencyCode()
	}

	if !resultCurrency.IsValid() {
		return ErrInvalidCurrencyCode()
	}

	if baseCurrency == resultCurrency {
		return ErrCurrenciesShouldDiffer()
	}

	return nil
}
