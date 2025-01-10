package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Record a visit
func RecordVisit(db *gorm.DB, linkID uuid.UUID, userAgent, ipAddress, referrer string) error {
	visit := &LinkVisit{
		LinkID:    linkID,
		UserAgent: &userAgent,
		IPAddress: &ipAddress,
		Referrer:  &referrer,
	}

	// Create visit record and update link visit count in a transaction
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(visit).Error; err != nil {
			return err
		}

		// Update link visit count and last visited time
		if err := tx.Model(&Link{}).
			Where("id = ?", linkID).
			Updates(map[string]interface{}{
				"visits":          gorm.Expr("visits + 1"),
				"last_visited_at": time.Now(),
			}).Error; err != nil {
			return err
		}

		return nil
	})
}
