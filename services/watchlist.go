package services

import (
	"context"

	"github.com/ODawah/Trading-Insights/models"

	"github.com/ODawah/Trading-Insights/repository"
)

type WatchListService interface {
	AddToWatchlist(ctx context.Context, userID uint, ticker string) error
	RemoveFromWatchlist(ctx context.Context, userID uint, ticker string) error
	GetWatchlist(ctx context.Context, userID uint) ([]string, error)
}

type watchListService struct {
	watchListRepo repository.WatchListRepository
}

func NewWatchListService(watchListRepo repository.WatchListRepository) WatchListService {
	return &watchListService{watchListRepo: watchListRepo}
}

func (s *watchListService) AddToWatchlist(ctx context.Context, userID uint, ticker string) error {
	watchItem := &models.WatchItem{
		UserID: userID,
		Ticker: ticker,
	}
	return s.watchListRepo.AddToWatchlist(ctx, watchItem)
}

func (s *watchListService) RemoveFromWatchlist(ctx context.Context, userID uint, ticker string) error {
	return s.watchListRepo.RemoveFromWatchlist(ctx, userID, ticker)
}

func (s *watchListService) GetWatchlist(ctx context.Context, userID uint) ([]string, error) {
	return s.watchListRepo.GetWatchlist(ctx, userID)
}
