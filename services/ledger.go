package services

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ODawah/Trading-Insights/models"
	"github.com/ODawah/Trading-Insights/repository"
	"gorm.io/datatypes"
)

type LedgerService interface {
	RecordExchange(ctx context.Context, req *ExchangeRequest) (string, error)
	ListEntries(ctx context.Context, userID uint, currency string, limit int) ([]models.UserLedgerEntry, error)
	GetTradeEntries(ctx context.Context, userID uint, tradeID string) ([]models.UserLedgerEntry, error)
}

type ledgerService struct {
	repo repository.LedgerRepository
}

func NewLedgerService(repo repository.LedgerRepository) LedgerService {
	return &ledgerService{repo: repo}
}

type ExchangeRequest struct {
	UserID      uint                   `json:"user_id"`
	TradeID     string                 `json:"trade_id,omitempty"`
	FromCurrency string                `json:"from_currency"`
	FromAmount   float64               `json:"from_amount"`
	ToCurrency   string                `json:"to_currency"`
	ToAmount     float64               `json:"to_amount"`
	FeeCurrency  string                `json:"fee_currency,omitempty"`
	FeeAmount    float64               `json:"fee_amount,omitempty"`
	ExecutedAt   time.Time             `json:"executed_at"`
	Meta         map[string]any        `json:"meta,omitempty"`
}

func (s *ledgerService) RecordExchange(ctx context.Context, req *ExchangeRequest) (string, error) {
	if err := validateExchangeRequest(req); err != nil {
		return "", err
	}

	tradeID := req.TradeID
	if strings.TrimSpace(tradeID) == "" {
		var err error
		tradeID, err = newTradeID()
		if err != nil {
			return "", fmt.Errorf("generate trade id: %w", err)
		}
	}

	executedAt := req.ExecutedAt
	if executedAt.IsZero() {
		executedAt = time.Now().UTC()
	}

	metaBytes, err := json.Marshal(req.Meta)
	if err != nil {
		return "", fmt.Errorf("marshal meta: %w", err)
	}

	entries := []models.UserLedgerEntry{
		{
			UserID:     req.UserID,
			TradeID:    tradeID,
			Currency:   strings.ToUpper(req.FromCurrency),
			Amount:     -req.FromAmount,
			ExecutedAt: executedAt,
			EntryType:  models.LedgerEntryExchange,
			Meta:       datatypes.JSON(metaBytes),
		},
		{
			UserID:     req.UserID,
			TradeID:    tradeID,
			Currency:   strings.ToUpper(req.ToCurrency),
			Amount:     req.ToAmount,
			ExecutedAt: executedAt,
			EntryType:  models.LedgerEntryExchange,
			Meta:       datatypes.JSON(metaBytes),
		},
	}

	if req.FeeAmount != 0 {
		feeCurrency := req.FeeCurrency
		if feeCurrency == "" {
			feeCurrency = req.FromCurrency
		}
		entries = append(entries, models.UserLedgerEntry{
			UserID:     req.UserID,
			TradeID:    tradeID,
			Currency:   strings.ToUpper(feeCurrency),
			Amount:     -req.FeeAmount,
			ExecutedAt: executedAt,
			EntryType:  models.LedgerEntryFee,
			Meta:       datatypes.JSON(metaBytes),
		})
	}

	if err := s.repo.Append(ctx, entries); err != nil {
		return "", err
	}
	return tradeID, nil
}

func (s *ledgerService) ListEntries(ctx context.Context, userID uint, currency string, limit int) ([]models.UserLedgerEntry, error) {
	currency = strings.ToUpper(strings.TrimSpace(currency))
	return s.repo.ListByUser(ctx, userID, currency, limit)
}

func (s *ledgerService) GetTradeEntries(ctx context.Context, userID uint, tradeID string) ([]models.UserLedgerEntry, error) {
	if strings.TrimSpace(tradeID) == "" {
		return nil, fmt.Errorf("trade_id is required")
	}
	return s.repo.GetByTradeID(ctx, userID, tradeID)
}

func validateExchangeRequest(req *ExchangeRequest) error {
	if req.UserID == 0 {
		return fmt.Errorf("user_id is required")
	}
	if strings.TrimSpace(req.FromCurrency) == "" || strings.TrimSpace(req.ToCurrency) == "" {
		return fmt.Errorf("from_currency and to_currency are required")
	}
	if req.FromAmount <= 0 || req.ToAmount <= 0 {
		return fmt.Errorf("amounts must be positive values")
	}
	return nil
}

// newTradeID generates a UUIDv4-like string without adding a dependency.
func newTradeID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80

	return fmt.Sprintf("%08x-%04x-%04x-%04x-%04x%08x",
		binary.BigEndian.Uint32(b[0:4]),
		binary.BigEndian.Uint16(b[4:6]),
		binary.BigEndian.Uint16(b[6:8]),
		binary.BigEndian.Uint16(b[8:10]),
		binary.BigEndian.Uint16(b[10:12]),
		binary.BigEndian.Uint32(b[12:16]),
	), nil
}
