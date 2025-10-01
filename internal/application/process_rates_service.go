package application

import (
	"context"
	"log/slog"
	"sync"

	"currency-rate-app/internal/common/utils"
	"currency-rate-app/internal/domains/currency"
	"currency-rate-app/internal/infrastructure/db"
	rates_api "currency-rate-app/internal/infrastructure/http/rates-api"
)

type ProcessRatesService struct {
	repo         db.CurrencyRepository
	ratesService rates_api.RateService
}

func NewProcessRatesService(
	repo db.CurrencyRepository,
	rateApi rates_api.RateService,
) *ProcessRatesService {
	return &ProcessRatesService{repo: repo, ratesService: rateApi}
}

type currencyPairGroup struct {
	ResultCurrency currency.CurrencyCode
	Ids            []string
}

func (s *ProcessRatesService) ProcessRates(ctx context.Context, batchSize int) {
	rates, err := s.repo.FetchAndMarkForProcessing(ctx, batchSize)

	if err != nil {
		slog.ErrorContext(
			ctx,
			"Rates fetch failed",
			slog.String("error", err.Error()),
		)

		return
	}

	if len(rates) == 0 {
		slog.InfoContext(ctx, "No rates to process")

		return
	}

	slog.InfoContext(ctx, "Processing rates")

	groupedRates := groupRates(rates)

	var wg sync.WaitGroup

	for baseCurrency, group := range groupedRates {
		b := baseCurrency
		g := group

		wg.Go(func() {
			s.processRatesGroup(ctx, b, g)
		})
	}

	wg.Wait()

	slog.InfoContext(ctx, "Finished processing rates")
}

func (s *ProcessRatesService) processRatesGroup(
	ctx context.Context,
	baseCurrency currency.CurrencyCode,
	group []currencyPairGroup,
) {
	defer utils.HandleRecover()

	rate, err := s.ratesService.FetchData(baseCurrency)

	if err != nil {
		slog.ErrorContext(
			ctx,
			"Error getting rate:",
			slog.String("error", err.Error()),
		)

		return
	}

	for _, val := range group {
		pairRate, ok := rate[string(val.ResultCurrency)]

		if !ok {
			if queryErr := s.repo.UpdateRateStatusByIds(ctx, val.Ids, currency.CurrencyRateStatusFailed); queryErr != nil {
				slog.ErrorContext(ctx, "Update failed", slog.String("error", queryErr.Error()))
			}
			slog.ErrorContext(
				ctx,
				"Currency pair not found",
				slog.String("baseCurrency", string(baseCurrency)),
				slog.String("resultCurrency", string(val.ResultCurrency)),
			)

			continue
		}

		if queryErr := s.repo.SaveRatesByIds(ctx, val.Ids, pairRate); queryErr != nil {
			slog.ErrorContext(ctx, "Update failed", slog.String("error", queryErr.Error()))
		}
	}
}

func groupRates(rates []currency.CurrencyRate) map[currency.CurrencyCode][]currencyPairGroup {
	groupedTemp := make(map[currency.CurrencyCode]map[currency.CurrencyCode][]string)

	for _, r := range rates {
		if groupedTemp[r.BaseCurrency] == nil {
			groupedTemp[r.BaseCurrency] = make(map[currency.CurrencyCode][]string)
		}
		groupedTemp[r.BaseCurrency][r.ResultCurrency] = append(groupedTemp[r.BaseCurrency][r.ResultCurrency], r.Id)
	}

	grouped := make(map[currency.CurrencyCode][]currencyPairGroup)

	for base, m := range groupedTemp {
		for result, ids := range m {
			grouped[base] = append(grouped[base], currencyPairGroup{ResultCurrency: result, Ids: ids})
		}
	}

	return grouped

}

func flattenGroupIds(items []currencyPairGroup) []string {
	var merged []string
	for _, it := range items {
		merged = append(merged, it.Ids...)
	}
	return merged
}
