package models

import (
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SecretKey represents the secret_keys table
type SecretKey struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	Key        string    `gorm:"type:varchar(64);unique;not null"`
	Name       string    `gorm:"type:varchar(100);not null"`
	CreatedAt  time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt  time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
	LastUsedAt *time.Time
	Active     bool `gorm:"not null;default:true"`
	IsAdmin    bool `gorm:"not null;default:false"`
}

// Link represents the links table
type Link struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	RedirectTo    string    `gorm:"type:varchar(2048);not null"`
	Shortened     string    `gorm:"type:varchar(100);unique;not null"`
	ExpiresAt     *time.Time
	CreatedAt     time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt     time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
	CreatedBy     uuid.UUID `gorm:"type:uuid;not null"`                 // Reference the SecretKey's ID
	SecretKey     SecretKey `gorm:"foreignKey:CreatedBy;references:ID"` // Foreign key references SecretKey's ID
	Visits        int       `gorm:"not null;default:0"`
	LastVisitedAt *time.Time
	Active        bool `gorm:"not null;default:true"`
}

// LinkVisit represents the link_visits table
type LinkVisit struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	LinkID    uuid.UUID `gorm:"type:uuid;not null"`
	Link      Link      `gorm:"foreignKey:LinkID"`
	VisitedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UserAgent *string   `gorm:"type:text"`
	IPAddress *string   `gorm:"type:inet"`
	Referrer  *string   `gorm:"type:text"`
}

// Request represents the requests table
type Request struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	IPAddress   string    `gorm:"type:inet;not null"`
	RequestedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
	SecretKeyID *uuid.UUID
	SecretKey   *SecretKey `gorm:"foreignKey:SecretKeyID"`
}

// SetupDatabase initializes the database schema and indexes
func SetupDatabase(db *gorm.DB) error {
	// Enable UUID extension
	db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"")

	// Auto-migrate the schemas in the correct order
	err := db.AutoMigrate(
		&SecretKey{}, // Create the secret_keys table first
		&Link{},      // Then create the links table
		&LinkVisit{},
		&Request{},
	)
	if err != nil {
		return err
	}

	// Create indexes
	err = createIndexes(db)
	if err != nil {
		return err
	}

	log.Println("✔️  Connected to postgres database.")

	return nil
}

// createIndexes sets up the necessary indexes
func createIndexes(db *gorm.DB) error {
	// Links indexes
	db.Exec("CREATE INDEX IF NOT EXISTS idx_links_expires_at ON links(expires_at) WHERE active = true")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_links_shortened ON links(shortened) WHERE active = true")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_links_created_by ON links(created_by)")

	// Secret keys index
	db.Exec("CREATE INDEX IF NOT EXISTS idx_secret_keys_key ON secret_keys(key) WHERE active = true")

	// Requests indexes
	db.Exec("CREATE INDEX IF NOT EXISTS idx_requests_ip_address_requested_at ON requests(ip_address, requested_at DESC)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_requests_secret_key_id_requested_at ON requests(secret_key_id, requested_at DESC)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_requests_requested_at ON requests(requested_at DESC)")

	return nil
}
