package application

import (
	"context"
	"testing"
	"time"

	"currency-rate-app/internal/domains/currency"
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

	if len(savedRates) != 2 {
		t.Fatalf("Expected 2 saved rates, got %d", len(savedRates))
	}
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

	if len(failedIds) != 1 || failedIds[0] != "1" {
		t.Errorf("Expected failed ID [1], got %v", failedIds)
	}
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

	if len(grouped) != 2 {
		t.Fatalf("Expected 2 base currencies, got %d", len(grouped))
	}

	usdGroup, exists := grouped[currency.USD]
	if !exists {
		t.Fatal("USD group not found")
	}
	if len(usdGroup) != 2 {
		t.Fatalf("Expected 2 currency pairs for USD, got %d", len(usdGroup))
	}

	eurGroup, exists := grouped[currency.EUR]
	if !exists {
		t.Fatal("EUR group not found")
	}
	if len(eurGroup) != 1 {
		t.Fatalf("Expected 1 currency pair for EUR, got %d", len(eurGroup))
	}
}

func TestGroupRates_EmptySlice(t *testing.T) {
	rates := []currency.CurrencyRate{}
	grouped := groupRates(rates)

	if len(grouped) != 0 {
		t.Errorf("Expected empty map, got %d items", len(grouped))
	}
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

	expectedLength := 6
	if len(flattened) != expectedLength {
		t.Fatalf("Expected %d, got %d", expectedLength, len(flattened))
	}
}

func TestFlattenString_EmptySlice(t *testing.T) {
	items := []currencyPairGroup{}
	flattened := flattenGroupIds(items)

	if len(flattened) != 0 {
		t.Errorf("Expected empty slice, got %d items", len(flattened))
	}
}
