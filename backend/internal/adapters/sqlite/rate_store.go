package sqlite

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/example/currency-watcher/backend/internal/currency"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type rateRow struct {
	Base      string `gorm:"primaryKey"`
	Target    string `gorm:"primaryKey"`
	Rate      float64
	ExpiresAt time.Time `gorm:"index"`
}

// RateStore retains cached exchange rates in SQLite.
type RateStore struct {
	db *gorm.DB
}

// NewRateStore opens and migrates the cache database.
func NewRateStore(databasePath string) (*RateStore, error) {
	if err := os.MkdirAll(filepath.Dir(databasePath), 0o755); err != nil {
		return nil, fmt.Errorf("create rate cache directory: %w", err)
	}

	db, err := gorm.Open(sqlite.Open(databasePath), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open rate cache database: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("access rate cache database: %w", err)
	}
	sqlDB.SetMaxOpenConns(1)
	if err := db.AutoMigrate(&rateRow{}); err != nil {
		return nil, fmt.Errorf("migrate rate cache database: %w", err)
	}

	return &RateStore{db: db}, nil
}

// FindFresh returns non-expired rates for the requested base and targets.
func (s *RateStore) FindFresh(ctx context.Context, base string, targets []string, now time.Time) (map[string]currency.CachedRate, error) {
	var entries []rateRow
	if err := s.db.WithContext(ctx).
		Where("base = ? AND target IN ?", base, targets).
		Find(&entries).Error; err != nil {
		return nil, fmt.Errorf("read rate cache: %w", err)
	}

	rates := make(map[string]currency.CachedRate, len(targets))
	for _, entry := range entries {
		if entry.ExpiresAt.After(now) {
			rates[entry.Target] = currency.CachedRate{Rate: entry.Rate, ExpiresAt: entry.ExpiresAt}
		}
	}
	return rates, nil
}

// Upsert persists rates by base-target pair.
func (s *RateStore) Upsert(ctx context.Context, base string, rates map[string]currency.CachedRate) error {
	entries := make([]rateRow, 0, len(rates))
	for target, rate := range rates {
		entries = append(entries, rateRow{
			Base:      base,
			Target:    target,
			Rate:      rate.Rate,
			ExpiresAt: rate.ExpiresAt,
		})
	}
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "base"}, {Name: "target"}},
			DoUpdates: clause.AssignmentColumns([]string{"rate", "expires_at"}),
		}).Create(&entries).Error
	}); err != nil {
		return fmt.Errorf("write rate cache: %w", err)
	}
	return nil
}

var _ currency.RateStore = (*RateStore)(nil)
