package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/ODawah/Trading-Insights/services"
	"github.com/go-chi/chi/v5"
)

type CurrenciesHandler struct {
	currencyService services.CurrencyService
}

func NewCurrenciesHandler(currencyService services.CurrencyService) *CurrenciesHandler {
	return &CurrenciesHandler{
		currencyService: currencyService,
	}
}

func (h *CurrenciesHandler) GetAllCurrenciesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		snapshot, err := h.currencyService.FetchLatestRates(r.Context())
		if err != nil {
			http.Error(w, "Failed to get cached snapshot", http.StatusInternalServerError)
			return
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

func (h *CurrenciesHandler) GetHistoryHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ticker := chi.URLParam(r, "ticker")
		query := r.URL.Query()

		var (
			fromPtr *time.Time
			toPtr   *time.Time
		)

		if from := query.Get("from"); from != "" {
			parsed, err := time.Parse(time.RFC3339, from)
			if err != nil {
				http.Error(w, "invalid 'from' timestamp; expected RFC3339", http.StatusBadRequest)
				return
			}
			fromPtr = &parsed
		}

		if to := query.Get("to"); to != "" {
			parsed, err := time.Parse(time.RFC3339, to)
			if err != nil {
				http.Error(w, "invalid 'to' timestamp; expected RFC3339", http.StatusBadRequest)
				return
			}
			toPtr = &parsed
		}

		limit := 500
		if raw := query.Get("limit"); raw != "" {
			if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
				limit = parsed
			}
		}

		rows, err := h.currencyService.FetchHistory(r.Context(), ticker, fromPtr, toPtr, limit)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(rows); err != nil {
			http.Error(w, "Failed to encode history", http.StatusInternalServerError)
			return
		}
	}
}

func (h *CurrenciesHandler) GetCandlesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ticker := chi.URLParam(r, "ticker")
		query := r.URL.Query()

		var (
			fromPtr *time.Time
			toPtr   *time.Time
		)

		if from := query.Get("from"); from != "" {
			parsed, err := time.Parse(time.RFC3339, from)
			if err != nil {
				http.Error(w, "invalid 'from' timestamp; expected RFC3339", http.StatusBadRequest)
				return
			}
			fromPtr = &parsed
		}

		if to := query.Get("to"); to != "" {
			parsed, err := time.Parse(time.RFC3339, to)
			if err != nil {
				http.Error(w, "invalid 'to' timestamp; expected RFC3339", http.StatusBadRequest)
				return
			}
			toPtr = &parsed
		}

		limit := 500
		if raw := query.Get("limit"); raw != "" {
			if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
				limit = parsed
			}
		}

		bucket := query.Get("bucket") // 1m,5m,15m,30m,1h,4h,1d
		rows, err := h.currencyService.FetchCandles(r.Context(), ticker, bucket, fromPtr, toPtr, limit)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(rows); err != nil {
			http.Error(w, "Failed to encode candles", http.StatusInternalServerError)
			return
		}
	}
}
