package persistence

import (
	"github.com/ODawah/Trading-Insights/models"
	"gorm.io/gorm"
)

type CurrenciesRepository interface {
	GetAllCurrencies() ([]models.Currencies, error)
	StoreCurrency(currency *models.Currency) error
	GetCurrencyByTicker(ticker string) (*models.Currency, error)
}

type GormCurrenciesRepo struct {
	db *gorm.DB
}

func NewCurrenciesRepository(db *gorm.DB) CurrenciesRepository {
	return &GormCurrenciesRepo{db: db}
}

func (c *GormCurrenciesRepo) GetAllCurrencies() ([]models.Currencies, error) {
	return []models.Currencies{}, nil
}

func (c *GormCurrenciesRepo) StoreCurrency(currency *models.Currency) error {
	return nil
}

func (c *GormCurrenciesRepo) GetCurrencyByTicker(ticker string) (*models.Currency, error) {
	return nil, nil
}
