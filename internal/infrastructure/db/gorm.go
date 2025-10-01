package db

import (
	"context"
	"currency-rate-app/internal/common/config"
	"fmt"
	"log/slog"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func SetupGorm(cfg *config.Config) *gorm.DB {
	databaseUrl := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s",
		cfg.DatabaseUsername,
		cfg.DatabasePassword,
		cfg.DatabaseHost,
		cfg.DatabasePort,
		cfg.DatabaseName,
	)
	db, err := gorm.Open(postgres.Open(databaseUrl), &gorm.Config{
		Logger: SlogLogger{},
	})

	if err != nil {
		panic(err)
	}

	err = db.AutoMigrate(&CurrencyRateEntity{})

	if err != nil {
		panic(err)
	}

	return db
}

type SlogLogger struct{}

func (l SlogLogger) LogMode(logger.LogLevel) logger.Interface { return l }

func (l SlogLogger) Info(ctx context.Context, msg string, _ ...interface{}) {
	slog.InfoContext(ctx, msg)
}

func (l SlogLogger) Warn(ctx context.Context, msg string, _ ...interface{}) {
	slog.WarnContext(ctx, msg)
}

func (l SlogLogger) Error(ctx context.Context, msg string, _ ...interface{}) {
	slog.ErrorContext(ctx, msg)
}

func (l SlogLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	sql, rows := fc()

	if err != nil {
		slog.ErrorContext(ctx, "SQL failed",
			slog.String("sql", sql),
			slog.Int64("rows", rows),
			slog.Int64("duration_ms", time.Since(begin).Milliseconds()),
			slog.String("error", err.Error()),
		)
		return
	}
	slog.InfoContext(ctx, "SQL executed",
		slog.String("sql", sql),
		slog.Int64("rows", rows),
		slog.Int64("duration_ms", time.Since(begin).Milliseconds()),
	)
}
