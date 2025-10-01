package config

import (
	"currency-rate-app/internal/common/validation"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	// App
	Port                             int         `env:"PORT" validate:"required"`
	Environment                      Environment `env:"ENVIRONMENT" validate:"required"`
	GracefulShutdownTimeoutInSeconds int         `env:"GRACEFUL_SHUTDOWN_TIMEOUT_IN_SECONDS" env-default:"5"`

	// Cron job
	RatesUpdateCronInSeconds int `env:"RATES_UPDATE_CRON_IN_SECONDS" validate:"required,min=1,max=60000"`
	RatesUpdateBatchSize     int `env:"RATES_UPDATE_BATCH_SIZE" validate:"required,min=1,max=100"`

	// Http Clients
	HttpClientsDefaultTimeoutInSeconds int `env:"HTTP_CLIENTS_DEFAULT_TIMEOUT_IN_SECONDS"`

	// Database
	DatabaseHost     string `env:"DATABASE_HOST" validate:"required"`
	DatabaseUsername string `env:"DATABASE_USERNAME" validate:"required"`
	DatabasePassword string `env:"DATABASE_PASSWORD" validate:"required"`
	DatabaseName     string `env:"DATABASE_NAME" validate:"required"`
	DatabasePort     uint16 `env:"DATABASE_PORT" validate:"required"`

	// Rates API
	RatesApiType      string `env:"RATES_API_TYPE" validate:"required"`
	FrankfurterApiURL string `env:"FRANKFURTER_API_URL" validate:"required"`
}

func Load() *Config {
	cfg, err := readEnv()

	if err != nil {
		panic(err)
	}

	err = validation.GetValidator().Struct(cfg)

	if err != nil {
		panic(err)
	}

	if !cfg.Environment.IsValid() {
		panic("invalid environment value:" + cfg.Environment)
	}

	return &cfg
}

func readEnv() (Config, error) {
	var cfg Config
	var err error

	if _, statErr := os.Stat(".env"); statErr == nil {
		err = cleanenv.ReadConfig(".env", &cfg)
	} else {
		err = cleanenv.ReadEnv(&cfg)
	}

	return cfg, err
}
