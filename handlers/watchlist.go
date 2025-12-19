package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/ODawah/Trading-Insights/middleware"
	"github.com/ODawah/Trading-Insights/services"
)

type WatchListHandler struct {
	watchListService services.WatchListService
}

func NewWatchListHandler(watchListService services.WatchListService) *WatchListHandler {
	return &WatchListHandler{
		watchListService: watchListService,
	}
}

func (h *WatchListHandler) AddToWatchlist(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req struct {
		Ticker string `json:"ticker"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.Ticker == "" {
		http.Error(w, "Ticker is required", http.StatusBadRequest)
		return
	}

	claims, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err = h.watchListService.AddToWatchlist(ctx, claims.UserID, req.Ticker)
	if err != nil {
		http.Error(w, "Failed to add to watchlist", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *WatchListHandler) RemoveFromWatchlist(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req struct {
		Ticker string `json:"ticker"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.Ticker == "" {
		http.Error(w, "Ticker is required", http.StatusBadRequest)
		return
	}

	claims, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err = h.watchListService.RemoveFromWatchlist(ctx, claims.UserID, req.Ticker)
	if err != nil {
		http.Error(w, "Failed to remove from watchlist", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *WatchListHandler) GetWatchlist(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	claims, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	tickers, err := h.watchListService.GetWatchlist(ctx, claims.UserID)
	if err != nil {
		http.Error(w, "Failed to get watchlist", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(struct {
		Tickers []string `json:"tickers"`
	}{
		Tickers: tickers,
	})
}
