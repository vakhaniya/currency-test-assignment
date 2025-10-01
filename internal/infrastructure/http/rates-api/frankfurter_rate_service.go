package rates_api

import (
	"encoding/json"
	"net/http"
	"net/url"

	error_utils "currency-rate-app/internal/common/error-utils"
	"currency-rate-app/internal/domains/currency"
)

type getRatesResponse struct {
	Base  string             `json:"base"`
	Date  string             `json:"date"`
	Rates map[string]float64 `json:"rates"`
}

type FrankfurterRateService struct {
	httpClient *http.Client
	baseUrl    string
}

func NewFrankfurterRateService(
	httpClient *http.Client,
	baseUrl string,
) *FrankfurterRateService {
	return &FrankfurterRateService{
		httpClient: httpClient,
		baseUrl:    baseUrl,
	}
}

func (s *FrankfurterRateService) FetchData(baseCurrency currency.CurrencyCode) (map[string]float64, error) {
	endpointUrl := s.baseUrl + "/v1/latest"
	baseURL, _ := url.Parse(endpointUrl)
	params := url.Values{}
	params.Add("base", string(baseCurrency))
	baseURL.RawQuery = params.Encode()

	res, err := s.httpClient.Get(baseURL.String())

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, error_utils.ErrInternalServerError("frankfurter response with code" + res.Status)
	}

	var body getRatesResponse

	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		return nil, error_utils.ErrInternalServerError(err.Error())
	}

	return body.Rates, nil
}
