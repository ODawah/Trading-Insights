package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/ODawah/Trading-Insights/models"
	"gorm.io/gorm"
)

type LedgerRepository interface {
	Append(ctx context.Context, entries []models.UserLedgerEntry) error
	GetByTradeID(ctx context.Context, userID uint, tradeID string) ([]models.UserLedgerEntry, error)
	ListByUser(ctx context.Context, userID uint, currency string, limit int) ([]models.UserLedgerEntry, error)
	ListByUserBetween(ctx context.Context, userID uint, from, to time.Time) ([]models.UserLedgerEntry, error)
	BalancesBefore(ctx context.Context, userID uint, before time.Time) (map[string]float64, error)
}

type ledgerRepository struct {
	db *gorm.DB
}

func NewLedgerRepository(db *gorm.DB) LedgerRepository {
	return &ledgerRepository{db: db}
}

func (r *ledgerRepository) Append(ctx context.Context, entries []models.UserLedgerEntry) error {
	if len(entries) == 0 {
		return nil
	}
	if err := r.db.WithContext(ctx).Create(&entries).Error; err != nil {
		return fmt.Errorf("append ledger entries: %w", err)
	}
	return nil
}

func (r *ledgerRepository) GetByTradeID(ctx context.Context, userID uint, tradeID string) ([]models.UserLedgerEntry, error) {
	query := r.db.WithContext(ctx).
		Where("trade_id = ?", tradeID)

	if userID != 0 {
		query = query.Where("user_id = ?", userID)
	}

	var rows []models.UserLedgerEntry
	if err := query.
		Order("executed_at DESC, id DESC").
		Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("get ledger entries for trade %s: %w", tradeID, err)
	}
	return rows, nil
}

func (r *ledgerRepository) ListByUser(ctx context.Context, userID uint, currency string, limit int) ([]models.UserLedgerEntry, error) {
	if limit <= 0 || limit > 2000 {
		limit = 200
	}
	query := r.db.WithContext(ctx).
		Where("user_id = ?", userID)

	if currency != "" {
		query = query.Where("currency = ?", currency)
	}

	var rows []models.UserLedgerEntry
	if err := query.
		Order("executed_at DESC, id DESC").
		Limit(limit).
		Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("list ledger entries: %w", err)
	}
	return rows, nil
}

func (r *ledgerRepository) ListByUserBetween(ctx context.Context, userID uint, from, to time.Time) ([]models.UserLedgerEntry, error) {
	if userID == 0 {
		return nil, fmt.Errorf("user_id is required")
	}
	var rows []models.UserLedgerEntry
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Where("executed_at >= ?", from).
		Where("executed_at <= ?", to).
		Order("executed_at ASC, id ASC").
		Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("list ledger entries between: %w", err)
	}
	return rows, nil
}

func (r *ledgerRepository) BalancesBefore(ctx context.Context, userID uint, before time.Time) (map[string]float64, error) {
	if userID == 0 {
		return nil, fmt.Errorf("user_id is required")
	}

	type row struct {
		Currency string  `gorm:"column:currency"`
		Balance  float64 `gorm:"column:balance"`
	}
	var rows []row
	if err := r.db.WithContext(ctx).
		Raw(
			`SELECT currency, COALESCE(SUM(amount), 0) AS balance
			 FROM user_ledger_entries
			 WHERE user_id = ? AND executed_at < ?
			 GROUP BY currency`,
			userID,
			before,
		).
		Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("balances before: %w", err)
	}

	out := make(map[string]float64, len(rows))
	for _, rr := range rows {
		out[rr.Currency] = rr.Balance
	}
	return out, nil
}
