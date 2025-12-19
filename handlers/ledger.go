package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/ODawah/Trading-Insights/middleware"
	"github.com/ODawah/Trading-Insights/services"
	"github.com/go-chi/chi/v5"
)

type LedgerHandler struct {
	ledgerService services.LedgerService
}

func NewLedgerHandler(ledgerService services.LedgerService) *LedgerHandler {
	return &LedgerHandler{ledgerService: ledgerService}
}

// RecordExchange persists a trade as balanced ledger rows (outflow/inflow + optional fee).
func (h *LedgerHandler) RecordExchange(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req services.ExchangeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	req.UserID = claims.UserID // Enforce ownership from JWT, not body

	// Allow client to omit executed_at; defaulted in service if zero.
	if req.ExecutedAt.IsZero() {
		req.ExecutedAt = time.Now().UTC()
	}

	tradeID, err := h.ledgerService.RecordExchange(ctx, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"trade_id": tradeID,
	})
}

func (h *LedgerHandler) ListEntries(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	query := r.URL.Query()
	limit := 200
	if raw := query.Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	entries, err := h.ledgerService.ListEntries(ctx, claims.UserID, query.Get("currency"), limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}

func (h *LedgerHandler) GetTrade(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	tradeID := chi.URLParam(r, "tradeID")
	entries, err := h.ledgerService.GetTradeEntries(ctx, claims.UserID, tradeID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}
