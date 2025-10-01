package application

import (
	"context"
	"errors"
	"testing"
	"time"

	error_utils "currency-rate-app/internal/common/error-utils"
	"currency-rate-app/internal/domains/currency"
)

type mockRepo struct {
	getActualFn               func(ctx context.Context, base currency.CurrencyCode, result currency.CurrencyCode) (*currency.CurrencyRate, error)
	getByIdFn                 func(ctx context.Context, id string) (*currency.CurrencyRate, error)
	createFn                  func(ctx context.Context, base currency.CurrencyCode, result currency.CurrencyCode, idem string) (*currency.CurrencyRate, error)
	updateRateStatusByIds     func(ctx context.Context, ids []string, status currency.CurrencyRateStatus) error
	saveRatesByIds            func(ctx context.Context, ids []string, rate float64) error
	fetchAndMarkForProcessing func(ctx context.Context, limit int) ([]currency.CurrencyRate, error)
}

func (m *mockRepo) GetActualRateByCurrency(ctx context.Context, base currency.CurrencyCode, result currency.CurrencyCode) (*currency.CurrencyRate, error) {
	return m.getActualFn(ctx, base, result)
}
func (m *mockRepo) GetRateById(ctx context.Context, id string) (*currency.CurrencyRate, error) {
	return m.getByIdFn(ctx, id)
}
func (m *mockRepo) CreateRate(ctx context.Context, base currency.CurrencyCode, result currency.CurrencyCode, idem string) (*currency.CurrencyRate, error) {
	return m.createFn(ctx, base, result, idem)
}
func (m *mockRepo) UpdateRateStatusByIds(ctx context.Context, ids []string, status currency.CurrencyRateStatus) error {
	return m.updateRateStatusByIds(ctx, ids, status)
}
func (m *mockRepo) SaveRatesByIds(ctx context.Context, ids []string, rate float64) error {
	return m.saveRatesByIds(ctx, ids, rate)
}
func (m *mockRepo) FetchAndMarkForProcessing(ctx context.Context, limit int) ([]currency.CurrencyRate, error) {
	return m.fetchAndMarkForProcessing(ctx, limit)
}

func TestCurrencyService_GetActualRate(t *testing.T) {
	ctx := context.Background()
	rate := 1.23
	completedAt := testTime()

	repo := &mockRepo{
		getActualFn: func(ctx context.Context, base currency.CurrencyCode, result currency.CurrencyCode) (*currency.CurrencyRate, error) {
			return &currency.CurrencyRate{BaseCurrency: base, ResultCurrency: result, Rate: &rate, CompletedAt: &completedAt, Status: currency.CurrencyRateStatusCompleted}, nil
		},
	}
	service := NewCurrencyService(repo)

	res, err := service.GetActualRate(ctx, currency.USD, currency.EUR)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.BaseCurrency != currency.USD || res.ResultCurrency != currency.EUR || res.Rate == nil || *res.Rate != rate {
		t.Fatalf("unexpected result: %+v", res)
	}
}

func TestCurrencyService_GetCompletedRateById(t *testing.T) {
	ctx := context.Background()
	rate := 2.5
	completedAt := testTime()

	tests := []struct {
		name      string
		model     *currency.CurrencyRate
		repoErr   error
		expectErr error
	}{
		{"completed", &currency.CurrencyRate{Status: currency.CurrencyRateStatusCompleted, Rate: &rate, CompletedAt: &completedAt}, nil, nil},
		{"pending", &currency.CurrencyRate{Status: currency.CurrencyRateStatusPending}, nil, currency.ErrCurrencyRateNotCompletedYet()},
		{"processing", &currency.CurrencyRate{Status: currency.CurrencyRateStatusProcessing}, nil, currency.ErrCurrencyRateNotCompletedYet()},
		{"failed", &currency.CurrencyRate{Status: currency.CurrencyRateStatusFailed}, nil, currency.ErrCurrencyRateFetchFailed()},
		{"unknown_status", &currency.CurrencyRate{Status: "SOMETHING"}, nil, error_utils.ErrInternalServerError("error")},
		{"repo_error", nil, errors.New("query_error"), errors.New("query_error")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepo{
				getByIdFn: func(ctx context.Context, id string) (*currency.CurrencyRate, error) {
					if tt.repoErr != nil {
						return nil, tt.repoErr
					}
					return tt.model, nil
				},
			}

			service := NewCurrencyService(repo)
			res, err := service.GetCompletedRateById(ctx, "id")

			if tt.expectErr != nil {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if res == nil || res.Rate == nil || *res.Rate != rate {
				t.Fatalf("unexpected res: %+v", res)
			}
		})
	}
}

func TestCurrencyService_CreateRate(t *testing.T) {
	ctx := context.Background()

	repo := &mockRepo{
		createFn: func(ctx context.Context, base currency.CurrencyCode, result currency.CurrencyCode, idem string) (*currency.CurrencyRate, error) {
			return &currency.CurrencyRate{Id: "qid", BaseCurrency: base, ResultCurrency: result, IdempotencyKey: idem, Status: currency.CurrencyRateStatusPending}, nil
		},
	}

	service := NewCurrencyService(repo)
	res, err := service.CreateRate(ctx, currency.USD, currency.MXN, "idem")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Id == "" || res.IdempotencyKey != "idem" || res.BaseCurrency != currency.USD || res.ResultCurrency != currency.MXN {
		t.Fatalf("unexpected result: %+v", res)
	}
}

func testTime() (t time.Time) {
	return time.Date(2024, 10, 2, 15, 4, 5, 0, time.UTC)
}
