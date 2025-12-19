package services

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ODawah/Trading-Insights/models"
	"github.com/ODawah/Trading-Insights/repository"
	"net/http"
	"os"
	"time"
)

type IngestionService interface {
	FetchRates(ctx context.Context) (*models.Snapshot, error)
}

type ingestionService struct {
	client       *http.Client
	currencyRepo repository.CurrencyRepository
}

func NewIngestionAPIClient(currencyRepo repository.CurrencyRepository) IngestionService {
	return &ingestionService{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		currencyRepo: currencyRepo,
	}
}

func (s *ingestionService) FetchRates(ctx context.Context) (*models.Snapshot, error) {
	snapshot, err := s.FetchFromEXCONVERT(ctx)
	if err != nil {
		return nil, err
	}
	err = s.currencyRepo.StoreSnapshotCache(ctx, snapshot)
	if err != nil {
		return nil, err
	}
	err = s.currencyRepo.StoreSnapShotPG(ctx, snapshot)
	if err != nil {
		return nil, err
	}
	return snapshot, nil
}

func (s *ingestionService) FetchFromEXCONVERT(ctx context.Context) (*models.Snapshot, error) {
	url := os.Getenv("EXCONVERT_URL")
	if url == "" {
		return nil, fmt.Errorf(" EXCONVERT_URL environment variable is not set")
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	resp, err := s.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data from EXCONVERT: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 response from EXCONVERT: %d", resp.StatusCode)
	}
	var snapshot models.Snapshot
	if err = json.NewDecoder(resp.Body).Decode(&snapshot); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	snapshot.Timestamp = time.Now()
	return &snapshot, nil
}
