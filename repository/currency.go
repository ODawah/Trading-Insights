package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ODawah/Trading-Insights/models"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type CurrencyRepository interface {
	GetSnapShotCache(ctx context.Context) (*models.Snapshot, error)
	StoreSnapshotCache(ctx context.Context, snapshot *models.Snapshot) error
	GetSnapShotPG(ctx context.Context) (*models.Snapshot, error)
	StoreSnapShotPG(ctx context.Context, snapshot *models.Snapshot) error
	ListHistory(ctx context.Context, ticker string, from, to *time.Time, limit int) ([]models.Currency, error)
	ListSnapshotTimes(ctx context.Context, from, to *time.Time, limit int) ([]time.Time, error)
	ListRatesAtTimes(ctx context.Context, tickers []string, times []time.Time) ([]models.Currency, error)
	LatestRatesAtOrBefore(ctx context.Context, tickers []string, at time.Time) ([]models.Currency, error)
	ListPairedRates(ctx context.Context, tickerA, tickerB string, from, to *time.Time, limit int) ([]PairedRateRow, error)
	ListCandles(ctx context.Context, ticker string, bucketInterval string, from, to *time.Time, limit int) ([]CandleRow, error)
}

type currencyRepository struct {
	redis *redis.Client
	db    *gorm.DB
}

func NewCurrencyRepository(redisClient *redis.Client, db *gorm.DB) CurrencyRepository {
	return &currencyRepository{
		redis: redisClient,
		db:    db,
	}
}

func (r *currencyRepository) GetSnapShotCache(ctx context.Context) (*models.Snapshot, error) {
	val, err := r.redis.Get(ctx, "Currency").Result()
	if err != nil {
		return nil, err
	}
	var snapshot models.Snapshot
	err = json.Unmarshal([]byte(val), &snapshot)
	if err != nil {
		return nil, err
	}
	return &snapshot, nil
}

func (r *currencyRepository) StoreSnapshotCache(ctx context.Context, snapshot *models.Snapshot) error {
	encoded, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}
	err = r.redis.Set(ctx, "Currency", encoded, 35*time.Second).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *currencyRepository) StoreSnapShotPG(ctx context.Context, snapshot *models.Snapshot) error {
	now := snapshot.Timestamp
	var rates []models.Currency
	for ticker, rate := range snapshot.Result {
		rates = append(rates, models.Currency{
			Ticker:      ticker,
			Base:        snapshot.Base,
			Rate:        rate,
			FetchedTime: now,
		})
	}

	if err := r.db.WithContext(ctx).Create(&rates).Error; err != nil {
		return fmt.Errorf("store snapshot in postgres: %w", err)
	}
	return nil
}

func (r *currencyRepository) GetSnapShotPG(ctx context.Context) (*models.Snapshot, error) {
	var latest time.Time
	if err := r.db.WithContext(ctx).
		Model(&models.Currency{}).
		Select("MAX(fetched_time)").
		Scan(&latest).Error; err != nil {
		return nil, fmt.Errorf("get latest snapshot timestamp: %w", err)
	}
	if latest.IsZero() {
		return nil, fmt.Errorf("no currency snapshots found")
	}

	var currencies []models.Currency
	if err := r.db.WithContext(ctx).
		Where("fetched_time = ?", latest).
		Find(&currencies).Error; err != nil {
		return nil, fmt.Errorf("get snapshot rows: %w", err)
	}
	if len(currencies) == 0 {
		return nil, fmt.Errorf("no currencies found for latest snapshot")
	}
	snapshot := &models.Snapshot{
		Base:      currencies[0].Base,
		Result:    make(map[string]float64),
		Timestamp: latest,
	}
	for _, currency := range currencies {
		snapshot.Result[currency.Ticker] = currency.Rate
	}
	return snapshot, nil
}

// ListHistory returns time-ordered rates for a single currency so callers can build candles/history.
func (r *currencyRepository) ListHistory(ctx context.Context, ticker string, from, to *time.Time, limit int) ([]models.Currency, error) {
	if limit <= 0 || limit > 2000 {
		limit = 500
	}
	query := r.db.WithContext(ctx).
		Where("ticker = ?", ticker)

	if from != nil {
		query = query.Where("fetched_time >= ?", *from)
	}
	if to != nil {
		query = query.Where("fetched_time <= ?", *to)
	}

	var rows []models.Currency
	if err := query.
		Order("fetched_time DESC, id DESC").
		Limit(limit).
		Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("list currency history: %w", err)
	}
	return rows, nil
}

func (r *currencyRepository) ListSnapshotTimes(ctx context.Context, from, to *time.Time, limit int) ([]time.Time, error) {
	if limit <= 0 || limit > 5000 {
		limit = 500
	}

	query := r.db.WithContext(ctx).Model(&models.Currency{})
	if from != nil {
		query = query.Where("fetched_time >= ?", *from)
	}
	if to != nil {
		query = query.Where("fetched_time <= ?", *to)
	}

	var times []time.Time
	if err := query.
		Distinct("fetched_time").
		Order("fetched_time ASC").
		Limit(limit).
		Pluck("fetched_time", &times).Error; err != nil {
		return nil, fmt.Errorf("list snapshot times: %w", err)
	}
	return times, nil
}

func (r *currencyRepository) ListRatesAtTimes(ctx context.Context, tickers []string, times []time.Time) ([]models.Currency, error) {
	if len(tickers) == 0 || len(times) == 0 {
		return nil, nil
	}

	var rows []models.Currency
	if err := r.db.WithContext(ctx).
		Where("ticker IN ?", tickers).
		Where("fetched_time IN ?", times).
		Order("fetched_time ASC, ticker ASC, id ASC").
		Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("list rates at times: %w", err)
	}
	return rows, nil
}

func (r *currencyRepository) LatestRatesAtOrBefore(ctx context.Context, tickers []string, at time.Time) ([]models.Currency, error) {
	if len(tickers) == 0 {
		return nil, nil
	}

	// DISTINCT ON is Postgres-specific and works well for point-in-time lookups.
	var rows []models.Currency
	if err := r.db.WithContext(ctx).
		Model(&models.Currency{}).
		Select("DISTINCT ON (ticker) *").
		Where("ticker IN ?", tickers).
		Where("fetched_time <= ?", at).
		Order("ticker, fetched_time DESC, id DESC").
		Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("latest rates at or before: %w", err)
	}
	return rows, nil
}

type PairedRateRow struct {
	Time  time.Time `gorm:"column:time"`
	RateA float64   `gorm:"column:rate_a"`
	RateB float64   `gorm:"column:rate_b"`
	Base  string    `gorm:"column:base"`
}

func (r *currencyRepository) ListPairedRates(ctx context.Context, tickerA, tickerB string, from, to *time.Time, limit int) ([]PairedRateRow, error) {
	if limit <= 0 || limit > 5000 {
		limit = 1000
	}

	// In this codebase we ingest rates with a single base (USD). The base currency itself
	// typically is not stored as a ticker row, so when either side is USD we synthesize
	// it as a constant 1.0 series aligned to the other ticker's timestamps.
	//
	// This keeps analytics working for any↔any pairs including USD.
	upperA := strings.ToUpper(strings.TrimSpace(tickerA))
	upperB := strings.ToUpper(strings.TrimSpace(tickerB))

	query := ""
	args := []any{}
	timeCol := ""

	switch {
	case upperA == "USD" && upperB != "USD":
		query = `
			SELECT b.fetched_time AS time, 1.0 AS rate_a, b.rate AS rate_b, b.base AS base
			FROM currencies b
			WHERE b.ticker = ?
		`
		args = append(args, upperB)
		timeCol = "b.fetched_time"
	case upperB == "USD" && upperA != "USD":
		query = `
			SELECT a.fetched_time AS time, a.rate AS rate_a, 1.0 AS rate_b, a.base AS base
			FROM currencies a
			WHERE a.ticker = ?
		`
		args = append(args, upperA)
		timeCol = "a.fetched_time"
	default:
		query = `
			SELECT a.fetched_time AS time, a.rate AS rate_a, b.rate AS rate_b, a.base AS base
			FROM currencies a
			JOIN currencies b
			  ON a.fetched_time = b.fetched_time
			 AND a.base = b.base
			WHERE a.ticker = ? AND b.ticker = ?
		`
		args = append(args, upperA, upperB)
		timeCol = "a.fetched_time"
	}

	if from != nil {
		query += " AND " + timeCol + " >= ?"
		args = append(args, *from)
	}
	if to != nil {
		query += " AND " + timeCol + " <= ?"
		args = append(args, *to)
	}

	query += " ORDER BY " + timeCol + " ASC LIMIT ?"
	args = append(args, limit)

	var rows []PairedRateRow
	if err := r.db.WithContext(ctx).Raw(query, args...).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("list paired rates: %w", err)
	}
	return rows, nil
}

type CandleRow struct {
	Bucket  time.Time `gorm:"column:bucket" json:"bucket"`
	Open    float64   `gorm:"column:open" json:"open"`
	High    float64   `gorm:"column:high" json:"high"`
	Low     float64   `gorm:"column:low" json:"low"`
	Close   float64   `gorm:"column:close" json:"close"`
	Samples int64     `gorm:"column:samples" json:"samples"`
}

func (r *currencyRepository) ListCandles(ctx context.Context, ticker string, bucketInterval string, from, to *time.Time, limit int) ([]CandleRow, error) {
	if strings.TrimSpace(ticker) == "" {
		return nil, fmt.Errorf("ticker is required")
	}
	if strings.TrimSpace(bucketInterval) == "" {
		return nil, fmt.Errorf("bucket interval is required")
	}
	if limit <= 0 || limit > 5000 {
		limit = 500
	}

	query := `
		SELECT
			time_bucket(?::interval, fetched_time) AS bucket,
			first(rate, fetched_time) AS open,
			max(rate) AS high,
			min(rate) AS low,
			last(rate, fetched_time) AS close,
			count(*) AS samples
		FROM currencies
		WHERE ticker = ?
	`
	args := []any{bucketInterval, strings.ToUpper(strings.TrimSpace(ticker))}
	if from != nil {
		query += " AND fetched_time >= ?"
		args = append(args, *from)
	}
	if to != nil {
		query += " AND fetched_time <= ?"
		args = append(args, *to)
	}

	query += " GROUP BY bucket ORDER BY bucket ASC LIMIT ?"
	args = append(args, limit)

	var rows []CandleRow
	if err := r.db.WithContext(ctx).Raw(query, args...).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("list candles: %w", err)
	}
	return rows, nil
}
