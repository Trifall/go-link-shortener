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

type LinkExpirationWorker struct {
	db       *gorm.DB
	interval time.Duration
}

func NewLinkExpirationWorker(db *gorm.DB) *LinkExpirationWorker {
	return &LinkExpirationWorker{
		db:       db,
		interval: time.Minute,
	}
}

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

func (w *LinkExpirationWorker) processExpiredLinks() error {
	// Generate random prefix for expired links
	prefix := make([]byte, 12)
	if _, err := rand.Read(prefix); err != nil {
		return err
	}
	randomPrefix := "expired_" + base64.URLEncoding.EncodeToString(prefix)[:12] + "_"

	type Result struct {
		ID string
	}
	var results []Result

	// Use GORM transaction
	err := w.db.Transaction(func(tx *gorm.DB) error {
		// Find and update expired links
		result := tx.Model(&models.Link{}).
			Where("active = ? AND ("+
				"(expires_at IS NOT NULL AND expires_at < ?) OR "+
				"(last_visited_at IS NOT NULL AND last_visited_at + INTERVAL '90 days' < ?)"+
				")", true, time.Now(), time.Now()).
			Updates(map[string]interface{}{
				"active":     false,
				"shortened":  gorm.Expr("? || id::text", randomPrefix),
				"updated_at": time.Now(),
			}).
			Select("id").
			Find(&results)

		if result.Error != nil {
			return result.Error
		}

		return nil
	})

	if err != nil {
		return err
	}

	if len(results) > 0 {
		log.Printf("Processed %d expired links", len(results))
	}

	return nil
}
