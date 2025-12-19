package services

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/ODawah/Trading-Insights/repository"
)

type AnalyticsService interface {
	CrossRateHistory(ctx context.Context, currencyA, currencyB string, from, to *time.Time, limit int) ([]CrossPoint, error)
	Correlation(ctx context.Context, currencyA, currencyB string, from, to *time.Time, limit int) (*CorrelationResult, error)
	CrossRateAt(ctx context.Context, currencyA, currencyB string, at time.Time) (*CrossPoint, error)
	PortfolioValueAt(ctx context.Context, userID uint, inCurrency string, at time.Time) (*PortfolioValue, error)
	PortfolioValueHistory(ctx context.Context, userID uint, inCurrency string, from, to time.Time, limit int) ([]PortfolioValue, error)
}

type analyticsService struct {
	currencyRepo repository.CurrencyRepository
	ledgerRepo   repository.LedgerRepository
}

func NewAnalyticsService(currencyRepo repository.CurrencyRepository, ledgerRepo repository.LedgerRepository) AnalyticsService {
	return &analyticsService{
		currencyRepo: currencyRepo,
		ledgerRepo:   ledgerRepo,
	}
}

type CrossPoint struct {
	Time time.Time `json:"time"`
	Rate float64   `json:"rate"`
	Base string    `json:"base"`
	A    string    `json:"a"`
	B    string    `json:"b"`
}

func (s *analyticsService) CrossRateHistory(ctx context.Context, currencyA, currencyB string, from, to *time.Time, limit int) ([]CrossPoint, error) {
	a := strings.ToUpper(strings.TrimSpace(currencyA))
	b := strings.ToUpper(strings.TrimSpace(currencyB))
	if a == "" || b == "" {
		return nil, fmt.Errorf("currency_a and currency_b are required")
	}
	if a == b {
		return nil, fmt.Errorf("currencies must be different")
	}

	rows, err := s.currencyRepo.ListPairedRates(ctx, a, b, from, to, limit)
	if err != nil {
		return nil, err
	}

	out := make([]CrossPoint, 0, len(rows))
	for _, r := range rows {
		// Assumption: stored rates represent Base->Ticker (e.g. 1 USD = rate EUR).
		// Cross A in B: (Base->B) / (Base->A)
		if r.RateA == 0 {
			continue
		}
		out = append(out, CrossPoint{
			Time: r.Time,
			Rate: r.RateB / r.RateA,
			Base: r.Base,
			A:    a,
			B:    b,
		})
	}
	return out, nil
}

func (s *analyticsService) CrossRateAt(ctx context.Context, currencyA, currencyB string, at time.Time) (*CrossPoint, error) {
	from := &at
	to := &at
	points, err := s.CrossRateHistory(ctx, currencyA, currencyB, from, to, 1)
	if err != nil {
		return nil, err
	}
	if len(points) == 0 {
		return nil, fmt.Errorf("no rate available at requested time")
	}
	return &points[0], nil
}

type CorrelationResult struct {
	A           string    `json:"a"`
	B           string    `json:"b"`
	Base        string    `json:"base"`
	From        *time.Time `json:"from,omitempty"`
	To          *time.Time `json:"to,omitempty"`
	Samples     int       `json:"samples"`
	Correlation float64   `json:"correlation"`
}

func (s *analyticsService) Correlation(ctx context.Context, currencyA, currencyB string, from, to *time.Time, limit int) (*CorrelationResult, error) {
	a := strings.ToUpper(strings.TrimSpace(currencyA))
	b := strings.ToUpper(strings.TrimSpace(currencyB))
	if a == "" || b == "" {
		return nil, fmt.Errorf("currency_a and currency_b are required")
	}
	if a == b {
		return nil, fmt.Errorf("currencies must be different")
	}

	rows, err := s.currencyRepo.ListPairedRates(ctx, a, b, from, to, limit)
	if err != nil {
		return nil, err
	}
	if len(rows) < 3 {
		return nil, fmt.Errorf("not enough samples to compute correlation")
	}

	// Compute log-returns for each series.
	var (
		returnsA []float64
		returnsB []float64
		base     string
	)
	base = rows[0].Base

	for i := 1; i < len(rows); i++ {
		prev := rows[i-1]
		cur := rows[i]
		if prev.RateA <= 0 || prev.RateB <= 0 || cur.RateA <= 0 || cur.RateB <= 0 {
			continue
		}
		returnsA = append(returnsA, math.Log(cur.RateA/prev.RateA))
		returnsB = append(returnsB, math.Log(cur.RateB/prev.RateB))
	}
	if len(returnsA) < 2 {
		return nil, fmt.Errorf("not enough valid samples to compute correlation")
	}

	corr, err := pearsonCorrelation(returnsA, returnsB)
	if err != nil {
		return nil, err
	}

	return &CorrelationResult{
		A:           a,
		B:           b,
		Base:        base,
		From:        from,
		To:          to,
		Samples:     len(returnsA),
		Correlation: corr,
	}, nil
}

type PortfolioValue struct {
	Time  time.Time `json:"time"`
	In    string    `json:"in"`
	Value float64   `json:"value"`
}

func (s *analyticsService) PortfolioValueAt(ctx context.Context, userID uint, inCurrency string, at time.Time) (*PortfolioValue, error) {
	if userID == 0 {
		return nil, fmt.Errorf("user_id is required")
	}
	in := strings.ToUpper(strings.TrimSpace(inCurrency))
	if in == "" {
		in = "USD"
	}

	// Balances as-of 'at' (approximation: balances before at + entries between [at,at] are ignored).
	// For point-in-time valuation in this codebase, executed_at and fetched_time typically share the same timeline.
	balances, err := s.ledgerRepo.BalancesBefore(ctx, userID, at.Add(1*time.Nanosecond))
	if err != nil {
		return nil, err
	}

	tickers := make([]string, 0, len(balances)+1)
	for c := range balances {
		cc := strings.ToUpper(strings.TrimSpace(c))
		if cc != "" {
			tickers = append(tickers, cc)
		}
	}
	if in != "" {
		tickers = append(tickers, in)
	}
	tickers = uniqueStrings(tickers)

	rates, err := s.currencyRepo.LatestRatesAtOrBefore(ctx, tickers, at)
	if err != nil {
		return nil, err
	}

	rateByTicker := make(map[string]float64, len(rates))
	baseCurrency := ""
	for _, r := range rates {
		rateByTicker[strings.ToUpper(r.Ticker)] = r.Rate
		if baseCurrency == "" && strings.TrimSpace(r.Base) != "" {
			baseCurrency = strings.ToUpper(strings.TrimSpace(r.Base))
		}
	}
	if baseCurrency == "" {
		baseCurrency = "USD"
	}

	valueIn, err := valueBalancesIn(balances, rateByTicker, baseCurrency, in)
	if err != nil {
		return nil, err
	}

	return &PortfolioValue{
		Time:  at,
		In:    in,
		Value: valueIn,
	}, nil
}

func (s *analyticsService) PortfolioValueHistory(ctx context.Context, userID uint, inCurrency string, from, to time.Time, limit int) ([]PortfolioValue, error) {
	if userID == 0 {
		return nil, fmt.Errorf("user_id is required")
	}
	if to.Before(from) {
		return nil, fmt.Errorf("to must be >= from")
	}
	in := strings.ToUpper(strings.TrimSpace(inCurrency))
	if in == "" {
		in = "USD"
	}
	if limit <= 0 || limit > 2000 {
		limit = 500
	}

	startBalances, err := s.ledgerRepo.BalancesBefore(ctx, userID, from)
	if err != nil {
		return nil, err
	}
	entries, err := s.ledgerRepo.ListByUserBetween(ctx, userID, from, to)
	if err != nil {
		return nil, err
	}

	fromPtr := &from
	toPtr := &to
	times, err := s.currencyRepo.ListSnapshotTimes(ctx, fromPtr, toPtr, limit)
	if err != nil {
		return nil, err
	}
	if len(times) == 0 {
		return nil, fmt.Errorf("no price snapshots in requested time range")
	}

	// Determine which currencies matter (start balances + entries currencies + output currency).
	currencySet := make(map[string]struct{})
	for c := range startBalances {
		cc := strings.ToUpper(strings.TrimSpace(c))
		if cc != "" {
			currencySet[cc] = struct{}{}
		}
	}
	for _, e := range entries {
		cc := strings.ToUpper(strings.TrimSpace(e.Currency))
		if cc != "" {
			currencySet[cc] = struct{}{}
		}
	}
	if in != "" {
		currencySet[in] = struct{}{}
	}

	tickers := make([]string, 0, len(currencySet))
	for c := range currencySet {
		tickers = append(tickers, c)
	}
	sort.Strings(tickers)

	rows, err := s.currencyRepo.ListRatesAtTimes(ctx, tickers, times)
	if err != nil {
		return nil, err
	}

	// Map rates by time -> ticker -> rate/base.
	type rateInfo struct {
		rate float64
		base string
	}
	ratesByTime := make(map[time.Time]map[string]rateInfo)
	for _, r := range rows {
		t := r.FetchedTime
		if _, ok := ratesByTime[t]; !ok {
			ratesByTime[t] = make(map[string]rateInfo)
		}
		ratesByTime[t][strings.ToUpper(r.Ticker)] = rateInfo{
			rate: r.Rate,
			base: strings.ToUpper(strings.TrimSpace(r.Base)),
		}
	}

	// Iterate forward in time, apply ledger movements as they occur, and value using rates at each snapshot time.
	balances := make(map[string]float64, len(startBalances))
	for c, v := range startBalances {
		balances[strings.ToUpper(strings.TrimSpace(c))] = v
	}

	entryIdx := 0
	out := make([]PortfolioValue, 0, len(times))
	for _, t := range times {
		for entryIdx < len(entries) && !entries[entryIdx].ExecutedAt.After(t) {
			e := entries[entryIdx]
			cc := strings.ToUpper(strings.TrimSpace(e.Currency))
			balances[cc] += e.Amount
			entryIdx++
		}

		rateMap, ok := ratesByTime[t]
		if !ok {
			continue
		}
		baseCurrency := ""
		rateByTicker := make(map[string]float64, len(rateMap))
		for tick, info := range rateMap {
			rateByTicker[tick] = info.rate
			if baseCurrency == "" && info.base != "" {
				baseCurrency = info.base
			}
		}
		if baseCurrency == "" {
			baseCurrency = "USD"
		}

		valueIn, err := valueBalancesIn(balances, rateByTicker, baseCurrency, in)
		if err != nil {
			return nil, err
		}
		out = append(out, PortfolioValue{
			Time:  t,
			In:    in,
			Value: valueIn,
		})
	}

	return out, nil
}

func valueBalancesIn(balances map[string]float64, rateByTicker map[string]float64, baseCurrency, inCurrency string) (float64, error) {
	baseCurrency = strings.ToUpper(strings.TrimSpace(baseCurrency))
	inCurrency = strings.ToUpper(strings.TrimSpace(inCurrency))
	if baseCurrency == "" {
		return 0, fmt.Errorf("base currency is unknown")
	}
	if inCurrency == "" {
		inCurrency = baseCurrency
	}

	rateIn := 1.0
	if inCurrency != baseCurrency {
		r, ok := rateByTicker[inCurrency]
		if !ok || r <= 0 {
			return 0, fmt.Errorf("missing rate for output currency %s", inCurrency)
		}
		rateIn = r
	}

	totalBase := 0.0
	for currency, amount := range balances {
		c := strings.ToUpper(strings.TrimSpace(currency))
		if c == "" || amount == 0 {
			continue
		}
		if c == baseCurrency {
			totalBase += amount
			continue
		}
		r, ok := rateByTicker[c]
		if !ok || r <= 0 {
			// If we can't price a holding at the requested time, fail so callers see missing data.
			return 0, fmt.Errorf("missing rate for currency %s", c)
		}
		// rate is Base->Currency, so Currency->Base is 1/r.
		totalBase += amount / r
	}

	// Convert base to output currency: Base->In
	return totalBase * rateIn, nil
}

func uniqueStrings(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		ss := strings.ToUpper(strings.TrimSpace(s))
		if ss == "" {
			continue
		}
		if _, ok := seen[ss]; ok {
			continue
		}
		seen[ss] = struct{}{}
		out = append(out, ss)
	}
	sort.Strings(out)
	return out
}

func pearsonCorrelation(a, b []float64) (float64, error) {
	if len(a) != len(b) {
		return 0, fmt.Errorf("mismatched lengths")
	}
	n := len(a)
	if n < 2 {
		return 0, fmt.Errorf("not enough samples")
	}

	meanA := 0.0
	meanB := 0.0
	for i := 0; i < n; i++ {
		meanA += a[i]
		meanB += b[i]
	}
	meanA /= float64(n)
	meanB /= float64(n)

	var (
		num   float64
		denA  float64
		denB  float64
	)
	for i := 0; i < n; i++ {
		da := a[i] - meanA
		db := b[i] - meanB
		num += da * db
		denA += da * da
		denB += db * db
	}
	den := math.Sqrt(denA * denB)
	if den == 0 {
		return 0, fmt.Errorf("zero variance")
	}
	return num / den, nil
}
