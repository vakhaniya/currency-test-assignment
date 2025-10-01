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
	http_server "currency-rate-app/internal/common/http-server"
	"currency-rate-app/internal/domains/currency"

	"github.com/stretchr/testify/assert"
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

	assert.Equal(t, http.StatusOK, res.Code)
	var body GetCurrencyResponse
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("invalid response: %v", err)
	}

	assert.Equal(t, currency.USD, body.BaseCurrency)
	assert.Equal(t, currency.EUR, body.ResultCurrency)
	assert.Equal(t, rate, body.Rate)
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

	assert.Equal(t, http.StatusBadRequest, res.Code)
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

	assert.Equal(t, http.StatusBadRequest, res.Code)
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

	assert.Equal(t, http.StatusNotFound, res.Code)
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

	assert.Equal(t, http.StatusOK, res.Code)

	var body GetCurrencyResponse
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("invalid response: %v", err)
	}

	assert.Equal(t, currency.USD, body.BaseCurrency)
	assert.Equal(t, currency.MXN, body.ResultCurrency)
	assert.Equal(t, rate, body.Rate)
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

	assert.Equal(t, http.StatusNotFound, res.Code)
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

	assert.Equal(t, http.StatusOK, res.Code)
	var resDto CreateRateResponse
	if err := json.NewDecoder(res.Body).Decode(&resDto); err != nil {
		t.Fatalf("invalid response: %v", err)
	}
	assert.NotNil(t, resDto.Id)
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

	var resDto http_server.HttpErrorResponse
	if err := json.NewDecoder(res.Body).Decode(&resDto); err != nil {
		t.Fatalf("invalid response: %v", err)
	}

	assert.Equal(t, http.StatusBadRequest, res.Code)
	assert.Equal(t, "Idempotency key not defined", resDto.Message)
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

	var resDto http_server.HttpErrorResponse
	if err := json.NewDecoder(res.Body).Decode(&resDto); err != nil {
		t.Fatalf("invalid response: %v", err)
	}

	assert.Equal(t, http.StatusBadRequest, res.Code)
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

	var resDto http_server.HttpErrorResponse
	if err := json.NewDecoder(res.Body).Decode(&resDto); err != nil {
		t.Fatalf("invalid response: %v", err)
	}

	assert.Equal(t, http.StatusBadRequest, res.Code)
	assert.Equal(t, "CurrenciesShouldDiffer", resDto.Code)
}
