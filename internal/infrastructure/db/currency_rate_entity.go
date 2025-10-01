package db

import (
	"currency-rate-app/internal/domains/currency"
	"time"
)

type CurrencyRateEntity struct {
	Id             string `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	IdempotencyKey string `gorm:"uniqueIndex;not null"`
	BaseCurrency   string `gorm:"not null"`
	ResultCurrency string `gorm:"not null"`
	Status         string `gorm:"not null"`
	Rate           *float64
	CompletedAt    *time.Time
	CreatedAt      time.Time `gorm:"not null"`
	UpdatedAt      time.Time `gorm:"not null"`
}

func (CurrencyRateEntity) TableName() string {
	return "currencies_rates"
}

func toDomain(e *CurrencyRateEntity) *currency.CurrencyRate {
	return &currency.CurrencyRate{
		Id:             e.Id,
		IdempotencyKey: e.IdempotencyKey,
		BaseCurrency:   currency.CurrencyCode(e.BaseCurrency),
		ResultCurrency: currency.CurrencyCode(e.ResultCurrency),
		Status:         currency.CurrencyRateStatus(e.Status),
		Rate:           e.Rate,
		CompletedAt:    e.CompletedAt,
		CreatedAt:      e.CreatedAt,
		UpdatedAt:      e.UpdatedAt,
	}
}
