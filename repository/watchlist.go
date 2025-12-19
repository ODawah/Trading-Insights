package repository

import (
	"context"

	"github.com/ODawah/Trading-Insights/models"
	"gorm.io/gorm"
)

type WatchListRepository interface {
	AddToWatchlist(ctx context.Context, watchItem *models.WatchItem) error
	RemoveFromWatchlist(ctx context.Context, userID uint, ticker string) error
	GetWatchlist(ctx context.Context, userID uint) ([]string, error)
}

type watchListRepository struct {
	db *gorm.DB
}

func NewWatchListRepository(db *gorm.DB) WatchListRepository {
	return &watchListRepository{db: db}
}

func (r *watchListRepository) AddToWatchlist(ctx context.Context, watchItem *models.WatchItem) error {
	return r.db.WithContext(ctx).Create(watchItem).Error
}

func (r *watchListRepository) RemoveFromWatchlist(ctx context.Context, userID uint, ticker string) error {
	return r.db.WithContext(ctx).Where("user_id = ? AND ticker = ?", userID, ticker).Delete(&models.WatchItem{}).Error
}

func (r *watchListRepository) GetWatchlist(ctx context.Context, userID uint) ([]string, error) {
	var watchItems []models.WatchItem
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&watchItems).Error
	if err != nil {
		return nil, err
	}
	tickers := make([]string, len(watchItems))
	for i, item := range watchItems {
		tickers[i] = item.Ticker
	}
	return tickers, nil
}
