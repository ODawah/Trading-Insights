package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ODawah/Trading-Insights/middleware"
	"github.com/ODawah/Trading-Insights/services"
)

type AnalyticsHandler struct {
	analytics services.AnalyticsService
}

func NewAnalyticsHandler(analytics services.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{analytics: analytics}
}

func (h *AnalyticsHandler) CrossRate(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	a := query.Get("a")
	b := query.Get("b")

	from, to, err := parseOptionalFromTo(query.Get("from"), query.Get("to"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	limit := parseLimit(query.Get("limit"), 1000, 5000)
	points, err := h.analytics.CrossRateHistory(r.Context(), a, b, from, to, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(points)
}

// ChartCross returns a time series suitable for charting; optional bucket will compress data (e.g., 1d).
func (h *AnalyticsHandler) ChartCross(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	a := query.Get("a")
	b := query.Get("b")

	from, to, err := parseOptionalFromTo(query.Get("from"), query.Get("to"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	limit := parseLimit(query.Get("limit"), 2000, 5000)
	points, err := h.analytics.CrossRateHistory(r.Context(), a, b, from, to, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	bucketParam := query.Get("bucket")
	if bucketParam != "" {
		dur, err := parseBucketDuration(bucketParam)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		points = bucketCrossPoints(points, dur)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(points)
}

func (h *AnalyticsHandler) Correlation(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	a := query.Get("a")
	b := query.Get("b")

	from, to, err := parseOptionalFromTo(query.Get("from"), query.Get("to"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	limit := parseLimit(query.Get("limit"), 2000, 5000)
	res, err := h.analytics.Correlation(r.Context(), a, b, from, to, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (h *AnalyticsHandler) ConvertAt(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	fromCur := query.Get("from")
	toCur := query.Get("to")

	at := time.Now().UTC()
	if raw := query.Get("at"); raw != "" {
		parsed, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			http.Error(w, "invalid 'at' timestamp; expected RFC3339", http.StatusBadRequest)
			return
		}
		at = parsed
	}

	point, err := h.analytics.CrossRateAt(r.Context(), fromCur, toCur, at)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(point)
}

func (h *AnalyticsHandler) PortfolioValue(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	query := r.URL.Query()
	in := query.Get("in")

	at := time.Now().UTC()
	if raw := query.Get("at"); raw != "" {
		parsed, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			http.Error(w, "invalid 'at' timestamp; expected RFC3339", http.StatusBadRequest)
			return
		}
		at = parsed
	}

	res, err := h.analytics.PortfolioValueAt(ctx, claims.UserID, in, at)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (h *AnalyticsHandler) PortfolioHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	query := r.URL.Query()
	in := query.Get("in")

	fromRaw := query.Get("from")
	toRaw := query.Get("to")
	if fromRaw == "" || toRaw == "" {
		http.Error(w, "'from' and 'to' are required (RFC3339)", http.StatusBadRequest)
		return
	}
	from, err := time.Parse(time.RFC3339, fromRaw)
	if err != nil {
		http.Error(w, "invalid 'from' timestamp; expected RFC3339", http.StatusBadRequest)
		return
	}
	to, err := time.Parse(time.RFC3339, toRaw)
	if err != nil {
		http.Error(w, "invalid 'to' timestamp; expected RFC3339", http.StatusBadRequest)
		return
	}

	limit := parseLimit(query.Get("limit"), 500, 2000)
	points, err := h.analytics.PortfolioValueHistory(ctx, claims.UserID, in, from, to, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(points)
}

func parseOptionalFromTo(fromRaw, toRaw string) (*time.Time, *time.Time, error) {
	var fromPtr *time.Time
	var toPtr *time.Time

	if fromRaw != "" {
		parsed, err := time.Parse(time.RFC3339, fromRaw)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid 'from' timestamp; expected RFC3339")
		}
		fromPtr = &parsed
	}
	if toRaw != "" {
		parsed, err := time.Parse(time.RFC3339, toRaw)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid 'to' timestamp; expected RFC3339")
		}
		toPtr = &parsed
	}

	return fromPtr, toPtr, nil
}

func parseLimit(raw string, def int, max int) int {
	limit := def
	if raw == "" {
		return limit
	}
	parsed, err := strconv.Atoi(raw)
	if err != nil || parsed <= 0 {
		return limit
	}
	if parsed > max {
		return max
	}
	return parsed
}

// parseBucketDuration supports friendly strings for chart compression.
func parseBucketDuration(raw string) (time.Duration, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "1m":
		return time.Minute, nil
	case "5m":
		return 5 * time.Minute, nil
	case "15m":
		return 15 * time.Minute, nil
	case "30m":
		return 30 * time.Minute, nil
	case "1h":
		return time.Hour, nil
	case "4h":
		return 4 * time.Hour, nil
	case "1d":
		return 24 * time.Hour, nil
	case "1w":
		return 7 * 24 * time.Hour, nil
	default:
		return 0, fmt.Errorf("unsupported bucket; use one of 1m,5m,15m,30m,1h,4h,1d,1w")
	}
}

// bucketCrossPoints keeps the last rate per bucket window (time ascending).
func bucketCrossPoints(points []services.CrossPoint, bucket time.Duration) []services.CrossPoint {
	if bucket <= 0 || len(points) == 0 {
		return points
	}
	var out []services.CrossPoint
	currentBucket := points[0].Time.Truncate(bucket)
	last := points[0]
	for _, p := range points {
		b := p.Time.Truncate(bucket)
		if b != currentBucket {
			out = append(out, last)
			currentBucket = b
		}
		last = p
	}
	// push final bucket
	out = append(out, last)
	return out
}
