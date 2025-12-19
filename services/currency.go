package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ODawah/Trading-Insights/models"
	"github.com/ODawah/Trading-Insights/repository"
)

type CurrencyService interface {
	FetchLatestRates(ctx context.Context) (*models.Snapshot, error)
	FetchHistory(ctx context.Context, ticker string, from, to *time.Time, limit int) ([]models.Currency, error)
	FetchCandles(ctx context.Context, ticker string, bucket string, from, to *time.Time, limit int) ([]repository.CandleRow, error)
}

type currencyService struct {
	currencyRepo repository.CurrencyRepository
}

func NewCurrencyService(currencyRepo repository.CurrencyRepository) CurrencyService {
	return &currencyService{
		currencyRepo: currencyRepo,
	}
}

func (s *currencyService) FetchLatestRates(ctx context.Context) (*models.Snapshot, error) {
	var snapshot *models.Snapshot
	snapshot, err := s.currencyRepo.GetSnapShotCache(ctx)
	if err != nil {
		snapshot, err = s.currencyRepo.GetSnapShotPG(ctx)
		if err != nil {
			return nil, err
		}

	}
	return snapshot, nil
}

func (s *currencyService) FetchHistory(ctx context.Context, ticker string, from, to *time.Time, limit int) ([]models.Currency, error) {
	if strings.TrimSpace(ticker) == "" {
		return nil, fmt.Errorf("ticker is required")
	}
	normalized := strings.ToUpper(strings.TrimSpace(ticker))
	return s.currencyRepo.ListHistory(ctx, normalized, from, to, limit)
}

func (s *currencyService) FetchCandles(ctx context.Context, ticker string, bucket string, from, to *time.Time, limit int) ([]repository.CandleRow, error) {
	if strings.TrimSpace(ticker) == "" {
		return nil, fmt.Errorf("ticker is required")
	}
	interval, err := normalizeCandleBucket(bucket)
	if err != nil {
		return nil, err
	}
	normalized := strings.ToUpper(strings.TrimSpace(ticker))
	return s.currencyRepo.ListCandles(ctx, normalized, interval, from, to, limit)
}

func normalizeCandleBucket(bucket string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(bucket)) {
	case "", "1m":
		return "1 minute", nil
	case "5m":
		return "5 minutes", nil
	case "15m":
		return "15 minutes", nil
	case "30m":
		return "30 minutes", nil
	case "1h":
		return "1 hour", nil
	case "4h":
		return "4 hours", nil
	case "1d":
		return "1 day", nil
	default:
		return "", fmt.Errorf("unsupported bucket; use one of: 1m, 5m, 15m, 30m, 1h, 4h, 1d")
	}
}
