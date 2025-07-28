package services

import (
	"context"
	"encoding/json"
	"github.com/ODawah/Trading-Insights/models"
	"github.com/ODawah/Trading-Insights/persistence"
	"gorm.io/gorm"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func FetchAllCurrencies(cache persistence.RedisRepository, db gorm.DB) (*models.Snapshot, error) {
	var snapshot models.Snapshot
	currencies := os.Getenv("EXCONVERT_URL")
	request, err := http.NewRequest(http.MethodGet, currencies, nil)
	now := time.Now()
	if err != nil {
		return nil, err
	}
	res, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	snapshot.Timestamp = now
	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(resBody, &snapshot)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	encoded, err := json.Marshal(snapshot)
	if err != nil {
		return nil, err
	}
	go func(s models.Snapshot) {
		var rates []models.Currency
		for ticker, rate := range s.Result {
			rates = append(rates, models.Currency{
				Ticker:      ticker,
				Rate:        rate,
				FetchedTime: now,
			})
		}
		if err = db.Create(&rates).Error; err != nil {
			log.Printf("Error storing currencies in database: %v", err)
		} else {
			log.Printf("Stored %d currencies in database", len(rates))
		}
	}(snapshot)
	if err = cache.Set(ctx, "currencies:latest", encoded); err != nil {
		return nil, err
	}
	return &snapshot, err
}

func GetCachedSnapshot(cache persistence.RedisRepository) (*models.Snapshot, error) {
	ctx := context.Background()
	raw, err := cache.Get(ctx, "currencies:latest")
	if err != nil {
		return nil, err
	}
	var snapshot models.Snapshot
	if err = json.Unmarshal([]byte(raw), &snapshot); err != nil {
		return nil, err
	}
	return &snapshot, nil
}
