package currency

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"currency-rate-app/internal/application"
	error_utils "currency-rate-app/internal/common/error-utils"
	"currency-rate-app/internal/domains/currency"
)

type mockService struct {
	getActualFn func(ctx context.Context, base currency.CurrencyCode, result currency.CurrencyCode) (*currency.CurrencyRate, error)
	getByIdFn   func(ctx context.Context, id string) (*currency.CurrencyRate, error)
	createFn    func(ctx context.Context, base currency.CurrencyCode, result currency.CurrencyCode, idem string) (*currency.CurrencyRate, error)
}

func (m *mockService) GetActualRate(ctx context.Context, base currency.CurrencyCode, result currency.CurrencyCode) (*currency.CurrencyRate, error) {
	return m.getActualFn(ctx, base, result)
}
func (m *mockService) GetCompletedRateById(ctx context.Context, id string) (*currency.CurrencyRate, error) {
	return m.getByIdFn(ctx, id)
}
func (m *mockService) CreateRate(ctx context.Context, base currency.CurrencyCode, result currency.CurrencyCode, idem string) (*currency.CurrencyRate, error) {
	return m.createFn(ctx, base, result, idem)
}

func fixedTime() time.Time { return time.Date(2024, 10, 2, 15, 4, 5, 0, time.UTC) }

func setupMux(currencyService application.CurrencyService) *http.ServeMux {
	mux := http.NewServeMux()
	NewCurrencyController(mux, currencyService)
	return mux
}

func TestGetActualCurrencyHandler_Success(t *testing.T) {
	rate := 1.11
	completed := fixedTime()
	currencyService := &mockService{
		getActualFn: func(ctx context.Context, base currency.CurrencyCode, result currency.CurrencyCode) (*currency.CurrencyRate, error) {
			return &currency.CurrencyRate{BaseCurrency: base, ResultCurrency: result, Status: currency.CurrencyRateStatusCompleted, Rate: &rate, CompletedAt: &completed}, nil
		},
	}
	mux := setupMux(currencyService)

	req := httptest.NewRequest(http.MethodGet, "/v1/currencies/actual?baseCurrency=USD&resultCurrency=EUR", nil)
	res := httptest.NewRecorder()

	mux.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.Code)
	}
	var body GetCurrencyResponse
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("invalid response: %v", err)
	}
	if body.BaseCurrency != currency.USD || body.ResultCurrency != currency.EUR || body.Rate != rate {
		t.Fatalf("unexpected body: %+v", body)
	}
}

func TestGetActualCurrencyHandler_WrongCurrencyCodeError(t *testing.T) {
	currencyService := &mockService{
		getActualFn: func(ctx context.Context, base currency.CurrencyCode, result currency.CurrencyCode) (*currency.CurrencyRate, error) {
			return nil, nil
		},
	}
	mux := setupMux(currencyService)

	req := httptest.NewRequest(http.MethodGet, "/v1/currencies/actual?baseCurrency=AAA&resultCurrency=EUR", nil)
	res := httptest.NewRecorder()

	mux.ServeHTTP(res, req)

	if res.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", res.Code)
	}
}

func TestGetActualCurrencyHandler_SameCurrencyCodesError(t *testing.T) {
	currencyService := &mockService{
		getActualFn: func(ctx context.Context, base currency.CurrencyCode, result currency.CurrencyCode) (*currency.CurrencyRate, error) {
			return nil, nil
		},
	}
	mux := setupMux(currencyService)

	req := httptest.NewRequest(http.MethodGet, "/v1/currencies/actual?baseCurrency=EUR&resultCurrency=EUR", nil)
	res := httptest.NewRecorder()

	mux.ServeHTTP(res, req)

	if res.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", res.Code)
	}
}

func TestGetActualCurrencyHandler_ServiceFail(t *testing.T) {
	currencyService := &mockService{
		getActualFn: func(ctx context.Context, base currency.CurrencyCode, result currency.CurrencyCode) (*currency.CurrencyRate, error) {
			return nil, &error_utils.CustomError{
				ErrorType: error_utils.ErrorCodeNotFound,
				Code:      "SomethingWrong",
			}
		},
	}
	mux := setupMux(currencyService)

	req := httptest.NewRequest(http.MethodGet, "/v1/currencies/actual?baseCurrency=MXN&resultCurrency=EUR", nil)
	res := httptest.NewRecorder()

	mux.ServeHTTP(res, req)

	if res.Code != http.StatusNotFound {
		t.Fatalf("expected 400, got %d", res.Code)
	}
}

func TestGetRateById_Success(t *testing.T) {
	rate := 2.22
	completed := fixedTime()
	currencyService := &mockService{
		getByIdFn: func(ctx context.Context, id string) (*currency.CurrencyRate, error) {
			return &currency.CurrencyRate{Id: id, BaseCurrency: currency.USD, ResultCurrency: currency.MXN, Status: currency.CurrencyRateStatusCompleted, Rate: &rate, CompletedAt: &completed}, nil
		},
	}
	mux := setupMux(currencyService)

	req := httptest.NewRequest(http.MethodGet, "/v1/currencies/id-1", nil)
	res := httptest.NewRecorder()

	mux.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.Code)
	}
	var body GetCurrencyResponse
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("invalid response: %v", err)
	}
	if body.BaseCurrency != currency.USD || body.ResultCurrency != currency.MXN || body.Rate != rate {
		t.Fatalf("unexpected body: %+v", body)
	}
}

func TestGetRateById_NotFound(t *testing.T) {
	currencyService := &mockService{
		getByIdFn: func(ctx context.Context, id string) (*currency.CurrencyRate, error) {
			return nil, currency.ErrCurrencyRateNotFound()
		},
	}
	mux := setupMux(currencyService)

	req := httptest.NewRequest(http.MethodGet, "/v1/currencies/unknown", nil)
	res := httptest.NewRecorder()

	mux.ServeHTTP(res, req)

	if res.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", res.Code)
	}
}

func TestCreateRate_Success(t *testing.T) {
	currencyService := &mockService{
		createFn: func(ctx context.Context, base currency.CurrencyCode, result currency.CurrencyCode, idem string) (*currency.CurrencyRate, error) {
			return &currency.CurrencyRate{Id: "qid-1", BaseCurrency: base, ResultCurrency: result, IdempotencyKey: idem, Status: currency.CurrencyRateStatusPending}, nil
		},
	}
	mux := setupMux(currencyService)

	dto := CreateRateRequest{BaseCurrency: currency.USD, ResultCurrency: currency.EUR}
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(dto)
	req := httptest.NewRequest(http.MethodPost, "/v1/currencies", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "idem-1")
	res := httptest.NewRecorder()

	mux.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.Code)
	}
	var resDto CreateRateResponse
	if err := json.NewDecoder(res.Body).Decode(&resDto); err != nil {
		t.Fatalf("invalid response: %v", err)
	}
	if resDto.Id == "" {
		t.Fatalf("expected id in response")
	}
}

func TestCreateRate_MissingHeader(t *testing.T) {
	currencyService := &mockService{
		createFn: func(ctx context.Context, base currency.CurrencyCode, result currency.CurrencyCode, idem string) (*currency.CurrencyRate, error) {
			return nil, nil
		},
	}
	mux := setupMux(currencyService)
	dto := CreateRateRequest{BaseCurrency: currency.USD, ResultCurrency: currency.EUR}
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(dto)
	req := httptest.NewRequest(http.MethodPost, "/v1/currencies", body)
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()

	mux.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.Code)
	}
}

func TestCreateRate_InvalidBody(t *testing.T) {
	currencyService := &mockService{
		createFn: func(ctx context.Context, base currency.CurrencyCode, result currency.CurrencyCode, idem string) (*currency.CurrencyRate, error) {
			return nil, nil
		},
	}
	mux := setupMux(currencyService)

	dto := CreateRateRequest{ResultCurrency: currency.EUR}
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(dto)
	req := httptest.NewRequest(http.MethodPost, "/v1/currencies", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "idem-1")
	res := httptest.NewRecorder()

	mux.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.Code)
	}
}

func TestCreateRate_SameCurrenciesError(t *testing.T) {
	currencyService := &mockService{
		createFn: func(ctx context.Context, base currency.CurrencyCode, result currency.CurrencyCode, idem string) (*currency.CurrencyRate, error) {
			return nil, nil
		},
	}
	mux := setupMux(currencyService)

	dto := CreateRateRequest{BaseCurrency: currency.EUR, ResultCurrency: currency.EUR}
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(dto)
	req := httptest.NewRequest(http.MethodPost, "/v1/currencies", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "idem-1")
	res := httptest.NewRecorder()

	mux.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.Code)
	}
}
