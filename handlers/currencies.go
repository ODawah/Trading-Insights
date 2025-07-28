package handlers

import (
	"encoding/json"
	"github.com/ODawah/Trading-Insights/persistence"
	"github.com/ODawah/Trading-Insights/services"
	"gorm.io/gorm"
	"net/http"
)

func GetAllCurrenciesHandler(redis persistence.RedisRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		snapshot, err := services.GetCachedSnapshot(redis)
		if err != nil {
			http.Error(w, "Failed to get cached snapshot", http.StatusInternalServerError)
			return
		}
		if snapshot == nil {
			snapshot, err = services.FetchAllCurrencies(redis, gorm.DB{})
			if err != nil {
				http.Error(w, "Failed to fetch all currencies", http.StatusInternalServerError)
				return
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(snapshot)
		if err != nil {
			http.Error(w, "Failed to encode snapshot", http.StatusInternalServerError)
			return
		}
	}
}
