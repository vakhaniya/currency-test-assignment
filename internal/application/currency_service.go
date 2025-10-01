package application

import (
	"context"
	error_utils "currency-rate-app/internal/common/error-utils"
	"currency-rate-app/internal/domains/currency"
	"currency-rate-app/internal/infrastructure/db"
)

type CurrencyService interface {
	GetActualRate(ctx context.Context, baseCurrency currency.CurrencyCode, resultCurrency currency.CurrencyCode) (*currency.CurrencyRate, error)
	GetCompletedRateById(ctx context.Context, id string) (*currency.CurrencyRate, error)
	CreateRate(ctx context.Context, baseCurrency currency.CurrencyCode, resultCurrency currency.CurrencyCode, idempotencyKey string) (*currency.CurrencyRate, error)
}

type currencyServiceImpl struct {
	repo db.CurrencyRepository
}

func NewCurrencyService(
	repo db.CurrencyRepository,
) CurrencyService {
	return &currencyServiceImpl{repo: repo}
}

func (s *currencyServiceImpl) GetActualRate(
	ctx context.Context,
	baseCurrency currency.CurrencyCode,
	resultCurrency currency.CurrencyCode,
) (*currency.CurrencyRate, error) {
	return s.repo.GetActualRateByCurrency(ctx, baseCurrency, resultCurrency)
}

func (s *currencyServiceImpl) GetCompletedRateById(
	ctx context.Context,
	id string,
) (*currency.CurrencyRate, error) {
	val, err := s.repo.GetRateById(ctx, id)

	if err != nil {
		return nil, err
	}

	if val.Status != currency.CurrencyRateStatusCompleted || val.Rate == nil {
		switch {
		case val.Status == currency.CurrencyRateStatusPending || val.Status == currency.CurrencyRateStatusProcessing:
			return nil, currency.ErrCurrencyRateNotCompletedYet()
		case val.Status == currency.CurrencyRateStatusFailed:
			return nil, currency.ErrCurrencyRateFetchFailed()
		default:
			return nil, error_utils.ErrInternalServerError("inconsistent entity state")
		}
	}

	return val, nil
}

func (s *currencyServiceImpl) CreateRate(
	ctx context.Context,
	baseCurrency currency.CurrencyCode,
	resultCurrency currency.CurrencyCode,
	idempotencyKey string,
) (*currency.CurrencyRate, error) {
	return s.repo.CreateRate(ctx, baseCurrency, resultCurrency, idempotencyKey)
}
