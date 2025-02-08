package workers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"go-link-shortener/models"
	"log"
	"time"

	"gorm.io/gorm"
)

// LinkExpirationWorker handles the scheduled expiration of links in the system
type LinkExpirationWorker struct {
	db       *gorm.DB
	interval time.Duration
}

// NewLinkExpirationWorker creates a new worker instance with the provided database connection
// The worker runs every 30 seconds by default
func NewLinkExpirationWorker(db *gorm.DB) *LinkExpirationWorker {
	return &LinkExpirationWorker{
		db:       db,
		interval: time.Minute / 2,
	}
}

// Start begins the worker process to handle link expirations
// It runs continuously until the provided context is cancelled
// Returns an error if the context is cancelled or if processing fails
func (w *LinkExpirationWorker) Start(ctx context.Context) error {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := w.processExpiredLinks(); err != nil {
				log.Printf("Error processing expired links: %v", err)
			}
		}
	}
}

// processExpiredLinks handles the deactivation of expired links
// Links are considered expired if:
// - Their expiration date has passed
// - They haven't been visited in 90 days
// Returns an error if database operations fail
func (w *LinkExpirationWorker) processExpiredLinks() error {
	prefix := make([]byte, 12)
	if _, err := rand.Read(prefix); err != nil {
		return err
	}
	randomPrefix := "expired_" + base64.URLEncoding.EncodeToString(prefix)[:12] + "_"

	var affectedRows int64
	var results []struct {
		ID string
	}

	err := w.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Link{}).
			Where("is_active = ? AND ("+
				"(expires_at IS NOT NULL AND expires_at < ?) OR "+
				"(last_visited_at IS NOT NULL AND last_visited_at + INTERVAL '90 days' < ?)"+
				")", true, time.Now(), time.Now()).
			Count(&affectedRows).Error; err != nil {
			return err
		}

		if affectedRows == 0 {
			return nil
		}

		return tx.Model(&models.Link{}).
			Where("is_active = ? AND ("+
				"(expires_at IS NOT NULL AND expires_at < ?) OR "+
				"(last_visited_at IS NOT NULL AND last_visited_at + INTERVAL '90 days' < ?)"+
				")", true, time.Now(), time.Now()).
			Updates(map[string]interface{}{
				"is_active":  false,
				"shortened":  gorm.Expr("? || id::text", randomPrefix),
				"updated_at": time.Now(),
			}).
			Scan(&results).Error
	})

	if err != nil {
		return err
	}

	if affectedRows > 0 {
		log.Printf("Processed %d expired links", affectedRows)
	}

	return nil
}
