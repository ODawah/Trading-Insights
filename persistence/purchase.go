package persistence

import "github.com/ODawah/Trading-Insights/models"

type purchaseRepository interface {
	Create(purchase *models.Purchase) error
	FindUserPurchases(userID uint) ([]models.Purchase, error)
	FindUserPurchasesByDate(userID uint, date string) ([]models.Purchase, error)
	FindUserPurchasesByTicker(userID uint, ticker string) ([]models.Purchase, error)
	TickerChangeRate(userID uint, ticker string, startDate string, endDate string) (float64, error)
}
