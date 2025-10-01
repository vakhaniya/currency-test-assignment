package application

import (
	"context"
	"testing"
	"time"

	"currency-rate-app/internal/domains/currency"

	"github.com/stretchr/testify/assert"
)

type mockCurrencyRepository struct {
	fetchAndMarkForProcessingFunc func(ctx context.Context, limit int) ([]currency.CurrencyRate, error)
	updateRateStatusByIdsFunc     func(ctx context.Context, ids []string, status currency.CurrencyRateStatus) error
	saveRatesByIdsFunc            func(ctx context.Context, ids []string, rate float64) error
}

func (m *mockCurrencyRepository) GetActualRateByCurrency(ctx context.Context, baseCurrency currency.CurrencyCode, resultCurrency currency.CurrencyCode) (*currency.CurrencyRate, error) {
	return nil, nil
}

func (m *mockCurrencyRepository) GetRateById(ctx context.Context, id string) (*currency.CurrencyRate, error) {
	return nil, nil
}

func (m *mockCurrencyRepository) CreateRate(ctx context.Context, baseCurrency currency.CurrencyCode, resultCurrency currency.CurrencyCode, idempotencyKey string) (*currency.CurrencyRate, error) {
	return nil, nil
}

func (m *mockCurrencyRepository) UpdateRateStatusByIds(ctx context.Context, ids []string, status currency.CurrencyRateStatus) error {
	if m.updateRateStatusByIdsFunc != nil {
		return m.updateRateStatusByIdsFunc(ctx, ids, status)
	}
	return nil
}

func (m *mockCurrencyRepository) SaveRatesByIds(ctx context.Context, ids []string, rate float64) error {
	if m.saveRatesByIdsFunc != nil {
		return m.saveRatesByIdsFunc(ctx, ids, rate)
	}
	return nil
}

func (m *mockCurrencyRepository) FetchAndMarkForProcessing(ctx context.Context, limit int) ([]currency.CurrencyRate, error) {
	if m.fetchAndMarkForProcessingFunc != nil {
		return m.fetchAndMarkForProcessingFunc(ctx, limit)
	}
	return nil, nil
}

type mockRateService struct {
	fetchDataFunc func(baseCurrency currency.CurrencyCode) (map[string]float64, error)
}

func (m *mockRateService) FetchData(baseCurrency currency.CurrencyCode) (map[string]float64, error) {
	if m.fetchDataFunc != nil {
		return m.fetchDataFunc(baseCurrency)
	}
	return nil, nil
}

func TestProcessRates_SuccessfulProcessing(t *testing.T) {
	now := time.Now()
	testRates := []currency.CurrencyRate{
		{
			Id:             "1",
			BaseCurrency:   currency.USD,
			ResultCurrency: currency.EUR,
			Status:         currency.CurrencyRateStatusProcessing,
			CreatedAt:      now,
			UpdatedAt:      now,
		},
		{
			Id:             "2",
			BaseCurrency:   currency.USD,
			ResultCurrency: currency.MXN,
			Status:         currency.CurrencyRateStatusProcessing,
			CreatedAt:      now,
			UpdatedAt:      now,
		},
	}

	var savedRates []struct {
		ids  []string
		rate float64
	}

	repo := &mockCurrencyRepository{
		fetchAndMarkForProcessingFunc: func(ctx context.Context, limit int) ([]currency.CurrencyRate, error) {
			return testRates, nil
		},
		saveRatesByIdsFunc: func(ctx context.Context, ids []string, rate float64) error {
			savedRates = append(savedRates, struct {
				ids  []string
				rate float64
			}{ids: ids, rate: rate})
			return nil
		},
	}

	rateService := &mockRateService{
		fetchDataFunc: func(baseCurrency currency.CurrencyCode) (map[string]float64, error) {
			return map[string]float64{
				"EUR": 0.85,
				"MXN": 20.5,
			}, nil
		},
	}

	service := NewProcessRatesService(repo, rateService)
	ctx := context.Background()

	service.ProcessRates(ctx, 10)

	expected := []struct {
		ids  []string
		rate float64
	}{
		{ids: []string{"1"}, rate: 0.85},
		{ids: []string{"2"}, rate: 20.5},
	}

	assert.Equal(t, expected, savedRates, "updated entities do not match")
}

func TestProcessRates_CurrencyPairNotFound(t *testing.T) {
	now := time.Now()
	testRates := []currency.CurrencyRate{
		{
			Id:             "1",
			BaseCurrency:   currency.USD,
			ResultCurrency: currency.EUR,
			Status:         currency.CurrencyRateStatusProcessing,
			CreatedAt:      now,
			UpdatedAt:      now,
		},
	}

	var failedIds []string

	repo := &mockCurrencyRepository{
		fetchAndMarkForProcessingFunc: func(ctx context.Context, limit int) ([]currency.CurrencyRate, error) {
			return testRates, nil
		},
		updateRateStatusByIdsFunc: func(ctx context.Context, ids []string, status currency.CurrencyRateStatus) error {
			failedIds = append(failedIds, ids...)
			return nil
		},
	}

	rateService := &mockRateService{
		fetchDataFunc: func(baseCurrency currency.CurrencyCode) (map[string]float64, error) {
			return map[string]float64{
				"MXN": 20.5,
			}, nil
		},
	}

	service := NewProcessRatesService(repo, rateService)
	ctx := context.Background()

	service.ProcessRates(ctx, 10)

	expected := []string{"1"}

	assert.Equal(t, expected, failedIds, "failed entities ids do not match")
}

func TestGroupRates(t *testing.T) {
	now := time.Now()
	rates := []currency.CurrencyRate{
		{
			Id:             "1",
			BaseCurrency:   currency.USD,
			ResultCurrency: currency.EUR,
			CreatedAt:      now,
		},
		{
			Id:             "2",
			BaseCurrency:   currency.USD,
			ResultCurrency: currency.MXN,
			CreatedAt:      now,
		},
		{
			Id:             "3",
			BaseCurrency:   currency.EUR,
			ResultCurrency: currency.USD,
			CreatedAt:      now,
		},
		{
			Id:             "4",
			BaseCurrency:   currency.USD,
			ResultCurrency: currency.EUR,
			CreatedAt:      now,
		},
	}

	grouped := groupRates(rates)

	expected := map[currency.CurrencyCode][]currencyPairGroup{
		currency.USD: {
			{Ids: []string{"1", "4"}, ResultCurrency: currency.EUR},
			{Ids: []string{"2"}, ResultCurrency: currency.MXN},
		},
		currency.EUR: {
			{Ids: []string{"3"}, ResultCurrency: currency.USD},
		},
	}

	assert.Equal(t, expected, grouped)
}

func TestGroupRates_EmptySlice(t *testing.T) {
	rates := []currency.CurrencyRate{}
	grouped := groupRates(rates)

	assert.Equal(t, len(grouped), 0, "expected empty map")
}

func TestFlattenString(t *testing.T) {
	items := []currencyPairGroup{
		{
			ResultCurrency: currency.EUR,
			Ids:            []string{"1", "2"},
		},
		{
			ResultCurrency: currency.MXN,
			Ids:            []string{"3"},
		},
		{
			ResultCurrency: currency.USD,
			Ids:            []string{"4", "5", "6"},
		},
	}

	flattened := flattenGroupIds(items)
	expected := []string{"1", "2", "3", "4", "5", "6"}
	assert.ElementsMatch(t, expected, flattened, "slices do not match")
}

func TestFlattenString_EmptySlice(t *testing.T) {
	items := []currencyPairGroup{}
	flattened := flattenGroupIds(items)

	assert.Equal(t, len(flattened), 0, "expected empty slice")
}
