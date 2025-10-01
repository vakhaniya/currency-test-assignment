package currency

import (
	"encoding/json"
	"net/http"

	"currency-rate-app/internal/application"
	error_utils "currency-rate-app/internal/common/error-utils"
	http_server "currency-rate-app/internal/common/http-server"
	"currency-rate-app/internal/common/validation"
	"currency-rate-app/internal/domains/currency"
)

type CurrencyController struct {
	service application.CurrencyService
}

func NewCurrencyController(
	mux *http.ServeMux,
	service application.CurrencyService,
) *CurrencyController {
	controller := &CurrencyController{service: service}

	mux.HandleFunc("GET /v1/currencies/actual", func(w http.ResponseWriter, r *http.Request) {
		controller.GetActualCurrencyHandler(w, r)
	})
	mux.HandleFunc("GET /v1/currencies/{id}", func(w http.ResponseWriter, r *http.Request) {
		controller.getCurrencyByIdHandler(w, r)
	})
	mux.HandleFunc("POST /v1/currencies", func(w http.ResponseWriter, r *http.Request) {
		controller.createCurrencyRateHandler(w, r)
	})

	return controller
}

// @Summary      Get actual currency rate
// @Description  Get actual currency rate by currency code
// @Tags         currency
// @Accept       json
// @Produce      json
// @Success 200  {object} GetCurrencyResponse
// @Param        baseCurrency   query      currency.CurrencyCode  true  "Currency Code"
// @Param        resultCurrency   query      currency.CurrencyCode  true  "Currency Code"
// @Router       /v1/currencies/actual [get]
func (c *CurrencyController) GetActualCurrencyHandler(w http.ResponseWriter, r *http.Request) {
	baseCurrency := currency.CurrencyCode(r.URL.Query().Get("baseCurrency"))
	resultCurrency := currency.CurrencyCode(r.URL.Query().Get("resultCurrency"))

	err := currency.ValidateCurrencyPair(baseCurrency, resultCurrency)

	if err != nil {
		http_server.SendErrorResponse(w, err)

		return
	}

	currencyRate, err := c.service.GetActualRate(r.Context(), baseCurrency, resultCurrency)

	if err != nil {
		http_server.SendErrorResponse(w, err)

		return
	}

	http_server.SendSuccessResponse(w, ToGetCurrencyResponse(*currencyRate))
}

// @Summary      Get currency rate by id
// @Description  Get currency rate by id
// @Tags         currency
// @Accept       json
// @Produce      json
// @Success 200  {object} GetCurrencyResponse
// @Param        id        path      string  true  "Id"
// @Router       /v1/currencies/{id} [get]
func (c *CurrencyController) getCurrencyByIdHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if id == "" {
		http_server.SendErrorResponse(w, currency.ErrCurrencyRateIdIsRequried())

		return
	}

	currencyRate, err := c.service.GetCompletedRateById(r.Context(), id)

	if err != nil {
		http_server.SendErrorResponse(w, err)

		return
	}

	http_server.SendSuccessResponse(w, ToGetCurrencyResponse(*currencyRate))
}

// @Summary      Create currency rate
// @Description  Create currency rate
// @Tags         currency
// @Accept       json
// @Produce      json
// @Success 200  {object} CreateRateResponse
// @Param 		 Idempotency-Key header string true "Idempotency key for request"
// @Param        request   body      CreateRateRequest  true  "Body"
// @Router       /v1/currencies [post]
func (c *CurrencyController) createCurrencyRateHandler(w http.ResponseWriter, r *http.Request) {
	var dto CreateRateRequest

	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http_server.SendErrorResponse(w, error_utils.ErrValidationError(err.Error()))

		return
	}

	if err := validation.GetValidator().Struct(&dto); err != nil {
		http_server.SendErrorResponse(w, error_utils.ErrValidationError(err.Error()))

		return
	}

	if err := currency.ValidateCurrencyPair(dto.BaseCurrency, dto.ResultCurrency); err != nil {
		http_server.SendErrorResponse(w, error_utils.ErrValidationError(err.Error()))

		return
	}

	idempotencyKey := r.Header.Get("Idempotency-Key")

	if idempotencyKey == "" {
		http_server.SendErrorResponse(w, error_utils.ErrValidationError("Idempotency key not defined"))

		return
	}

	currencyRate, err := c.service.CreateRate(r.Context(), dto.BaseCurrency, dto.ResultCurrency, idempotencyKey)

	if err != nil {
		http_server.SendErrorResponse(w, err)

		return
	}

	response := &CreateRateResponse{
		Id: currencyRate.Id,
	}

	http_server.SendSuccessResponse(w, response)
}
