package persistence

import "github.com/ODawah/Trading-Insights/models"

type WatchItemRepository interface {
	Create(watchItem *models.WatchItem) error
	FindWatchList(userID uint) ([]models.WatchItem, error)
	DeleteWatchItem(userID uint, ticker string) error
}
