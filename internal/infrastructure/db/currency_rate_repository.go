package db

import (
	"context"
	error_utils "currency-rate-app/internal/common/error-utils"
	"currency-rate-app/internal/domains/currency"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CurrencyRepository interface {
	GetActualRateByCurrency(ctx context.Context, baseCurrency currency.CurrencyCode, resultCurrency currency.CurrencyCode) (*currency.CurrencyRate, error)
	GetRateById(ctx context.Context, id string) (*currency.CurrencyRate, error)
	CreateRate(ctx context.Context, baseCurrency currency.CurrencyCode, resultCurrency currency.CurrencyCode, idempotencyKey string) (*currency.CurrencyRate, error)
	UpdateRateStatusByIds(ctx context.Context, ids []string, status currency.CurrencyRateStatus) error
	SaveRatesByIds(ctx context.Context, ids []string, rate float64) error
	FetchAndMarkForProcessing(ctx context.Context, limit int) ([]currency.CurrencyRate, error)
}

type currencyRepositoryImpl struct {
	db *gorm.DB
}

func NewCurrencyRepository(db *gorm.DB) *currencyRepositoryImpl {
	return &currencyRepositoryImpl{db: db}
}

func (repo *currencyRepositoryImpl) GetActualRateByCurrency(
	ctx context.Context,
	baseCurrency currency.CurrencyCode,
	resultCurrency currency.CurrencyCode,
) (*currency.CurrencyRate, error) {
	var task CurrencyRateEntity

	err := repo.db.
		WithContext(ctx).
		Where(&CurrencyRateEntity{
			BaseCurrency:   string(baseCurrency),
			ResultCurrency: string(resultCurrency),
			Status:         string(currency.CurrencyRateStatusCompleted),
		}).
		Order("completed_at DESC").
		First(&task).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, currency.ErrCurrencyRateNotFound()
		}

		return nil, error_utils.ErrInternalServerError(err.Error())
	}

	return toDomain(&task), nil
}

func (repo *currencyRepositoryImpl) GetRateById(ctx context.Context, id string) (*currency.CurrencyRate, error) {
	var task CurrencyRateEntity

	err := repo.db.WithContext(ctx).Where(&CurrencyRateEntity{Id: id}).First(&task).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, currency.ErrCurrencyRateNotFound()
		}

		return nil, error_utils.ErrInternalServerError(err.Error())
	}

	return toDomain(&task), nil
}

func (repo *currencyRepositoryImpl) CreateRate(
	ctx context.Context,
	baseCurrency currency.CurrencyCode,
	resultCurrency currency.CurrencyCode,
	idempotencyKey string,
) (*currency.CurrencyRate, error) {
	entity := &CurrencyRateEntity{
		BaseCurrency:   string(baseCurrency),
		ResultCurrency: string(resultCurrency),
		IdempotencyKey: idempotencyKey,
		Status:         string(currency.CurrencyRateStatusPending),
	}

	res := repo.db.WithContext(ctx).Clauses(
		clause.OnConflict{DoNothing: true, Columns: []clause.Column{{Name: "idempotency_key"}}},
	).Create(&entity)

	if res.Error != nil {
		return nil, error_utils.ErrInternalServerError(res.Error.Error())
	}

	if res.RowsAffected == 0 {
		var existingEntity CurrencyRateEntity
		err := repo.db.Where(&CurrencyRateEntity{IdempotencyKey: idempotencyKey}).First(&existingEntity).Error

		if err != nil {
			return nil, error_utils.ErrInternalServerError(err.Error())
		}

		if existingEntity.BaseCurrency != string(baseCurrency) || existingEntity.ResultCurrency != string(resultCurrency) {
			return nil, error_utils.ErrBusinessLogic("CurrencyRateIdempotencyConflict")
		}

		return toDomain(&existingEntity), nil
	}

	return toDomain(entity), nil
}

func (repo *currencyRepositoryImpl) UpdateRateStatusByIds(
	ctx context.Context,
	ids []string,
	status currency.CurrencyRateStatus,
) error {
	err := repo.db.WithContext(ctx).Where("id IN ?", ids).UpdateColumns(&CurrencyRateEntity{Status: string(status)}).Error

	return err
}

func (repo *currencyRepositoryImpl) SaveRatesByIds(ctx context.Context, ids []string, rate float64) error {
	now := time.Now()

	err := repo.db.WithContext(ctx).Where("id IN ?", ids).
		UpdateColumns(&CurrencyRateEntity{
			Status:      string(currency.CurrencyRateStatusCompleted),
			UpdatedAt:   now,
			CompletedAt: &now,
			Rate:        &rate,
		}).Error

	return err
}

func (repo *currencyRepositoryImpl) FetchAndMarkForProcessing(ctx context.Context, limit int) ([]currency.CurrencyRate, error) {
	var entities []CurrencyRateEntity

	err := repo.db.Transaction(func(tx *gorm.DB) error {
		subquery := tx.
			WithContext(ctx).
			Model(&CurrencyRateEntity{}).
			Where(&CurrencyRateEntity{Status: string(currency.CurrencyRateStatusPending)}).
			Order("created_at ASC").
			Limit(limit).
			Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Select("id")

		err := tx.Model(&entities).
			WithContext(ctx).
			Clauses(clause.Returning{}).
			Where("id IN (?)", subquery).
			UpdateColumns(&CurrencyRateEntity{
				Status:    string(currency.CurrencyRateStatusProcessing),
				UpdatedAt: time.Now(),
			}).Error

		return err
	})

	if err != nil {
		return nil, err
	}

	models := make([]currency.CurrencyRate, 0, len(entities))
	for _, task := range entities {
		models = append(models, *toDomain(&task))
	}
	return models, nil
}
