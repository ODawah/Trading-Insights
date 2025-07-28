package persistence

import (
	"github.com/ODawah/Trading-Insights/models"
	"gorm.io/gorm"
)

type CurrenciesRepository interface {
	GetAllCurrencies() ([]models.Snapshot, error)
	StoreCurrency(currency *models.Currency) error
	StoreCacheSnapshot(snapshot *models.Snapshot) error
	GetCurrencyByTicker(ticker string) (*models.Currency, error)
}

type GormCurrenciesRepo struct {
	db *gorm.DB
}

func NewCurrenciesRepository(db *gorm.DB) CurrenciesRepository {
	return &GormCurrenciesRepo{db: db}
}

func (c *GormCurrenciesRepo) GetAllCurrencies() ([]models.Snapshot, error) {
	return []models.Snapshot{}, nil
}

func (c *GormCurrenciesRepo) StoreCurrency(currency *models.Currency) error {
	return nil
}

func (c *GormCurrenciesRepo) GetCurrencyByTicker(ticker string) (*models.Currency, error) {
	return nil, nil
}

func (c *GormCurrenciesRepo) StoreCacheSnapshot(snapshot *models.Snapshot) error {
	if err := c.db.Create(snapshot).Error; err != nil {
		return err
	}
	return nil

}
