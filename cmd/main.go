package main

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"currency-rate-app/internal/api/currency"
	"currency-rate-app/internal/application"
	"currency-rate-app/internal/common/config"
	http_client "currency-rate-app/internal/common/http-client"
	"currency-rate-app/internal/common/logger"
	"currency-rate-app/internal/common/middlewares"
	"currency-rate-app/internal/common/utils"
	"currency-rate-app/internal/infrastructure/db"
	rateservice "currency-rate-app/internal/infrastructure/http/rates-api"

	_ "currency-rate-app/docs"

	httpSwagger "github.com/swaggo/http-swagger"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	logger.ConfigureLogger(cfg.Environment == config.Local)

	gorm := db.SetupGorm(cfg)
	serveMux := http.NewServeMux()

	server := setupServer(cfg, serveMux)
	cancel := setupApp(ctx, serveMux, cfg, gorm)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	utils.WaitForShutdown(ctx, server, cancel, cfg.GracefulShutdownTimeoutInSeconds)
}

func setupServer(cfg *config.Config, serveMux *http.ServeMux) *http.Server {
	middlewareStack := middlewares.ChainMiddlewares(
		middlewares.RecoveryMiddleware,
		middlewares.TracingMiddleware,
		middlewares.LoggingMiddleware,
	)

	server := http.Server{
		Addr:    ":" + strconv.Itoa(cfg.Port),
		Handler: middlewareStack(serveMux),
	}

	return &server
}

func setupApp(
	ctx context.Context,
	serveMux *http.ServeMux,
	cfg *config.Config,
	gorm *gorm.DB,
) context.CancelFunc {
	currencyRepoGorm := db.NewCurrencyRepository(gorm)
	currencyService := application.NewCurrencyService(currencyRepoGorm)

	currency.NewCurrencyController(serveMux, currencyService)
	serveMux.Handle("/swagger/", httpSwagger.WrapHandler)

	httpClient := &http.Client{
		Transport: http_client.LogRoundTripper{DefaultClient: http.DefaultTransport},
		Timeout:   time.Duration(cfg.HttpClientsDefaultTimeoutInSeconds) * time.Second,
	}

	rateApiService := rateservice.NewRateService(httpClient, *cfg)
	processRateService := application.NewProcessRatesService(currencyRepoGorm, rateApiService)

	cancel := utils.CreateCronJob(ctx, time.Duration(cfg.RatesUpdateCronInSeconds)*time.Second, func(cronCtx context.Context) {
		processRateService.ProcessRates(cronCtx, cfg.RatesUpdateBatchSize)
	})

	return cancel
}
