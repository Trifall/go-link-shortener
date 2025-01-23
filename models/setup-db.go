package models

import (
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SecretKey represents the secret_keys table
type SecretKey struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	Key       string    `gorm:"type:varchar(64);unique;not null" json:"key"`
	Name      string    `gorm:"type:varchar(100);not null" json:"name"`
	CreatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP;autoUpdateTime" json:"updated_at"`
	IsActive  bool      `gorm:"not null;default:true" json:"is_active"`
	IsAdmin   bool      `gorm:"not null;default:false" json:"is_admin"`
}

// Link represents the links table
type Link struct {
	ID            uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	RedirectTo    string     `gorm:"type:varchar(2048);not null" json:"redirect_to"`
	Shortened     string     `gorm:"type:varchar(100);unique;not null" json:"shortened"`
	ExpiresAt     *time.Time `json:"expires_at"`
	CreatedAt     time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP;autoUpdateTime" json:"updated_at"`
	CreatedBy     uuid.UUID  `gorm:"type:uuid;not null" json:"created_by"`
	SecretKey     SecretKey  `gorm:"foreignKey:CreatedBy;references:ID" json:"secret_key"`
	Visits        int        `gorm:"not null;default:0" json:"visits"`
	LastVisitedAt *time.Time `json:"last_visited_at"`
	IsActive      bool       `gorm:"not null;default:true" json:"is_active"`
}

// LinkVisit represents the link_visits table
type LinkVisit struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	LinkID    uuid.UUID `gorm:"type:uuid;not null" json:"link_id"`
	Link      Link      `gorm:"foreignKey:LinkID" json:"link"`
	VisitedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"visited_at"`
	UserAgent *string   `gorm:"type:text" json:"user_agent,omitempty"`
	IPAddress *string   `gorm:"type:inet" json:"ip_address,omitempty"`
	Referrer  *string   `gorm:"type:text" json:"referrer,omitempty"`
}

// Request represents the requests table
type Request struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	IPAddress   string    `gorm:"type:inet;not null" json:"ip_address"`
	RequestedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"requested_at"`
}

// LogType represents the type of log entry
type LogType string

const (
	LogTypeError   LogType = "error"
	LogTypeInfo    LogType = "info"
	LogTypeWarning LogType = "warning"
)

// LogSource represents the source of the log entry
type LogSource string

const (
	LogSourceDatabase LogSource = "database"
	LogSourceAuth     LogSource = "auth"
	LogSourceLinks    LogSource = "links"
	LogSourceRequest  LogSource = "request"
	LogSourceMisc     LogSource = "misc"
)

// Log represents the logs table
type Log struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	Timestamp time.Time `gorm:"not null;default:CURRENT_TIMESTAMP;index" json:"timestamp"`
	Type      LogType   `gorm:"type:varchar(10);not null;index" json:"type"`
	Source    LogSource `gorm:"type:varchar(20);not null;index" json:"source"`
	Message   string    `gorm:"type:text;not null" json:"message"`
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
		&Log{}, // Create the logs table
	)

	if err != nil {
		return err
	}

	// Create indexes
	err = createIndexes(db)
	if err != nil {
		return err
	}

	log.Println("✔️  Connected to Postgres database.")
	return nil
}

// createIndexes sets up the necessary indexes
func createIndexes(db *gorm.DB) error {
	// Links indexes
	db.Exec("CREATE INDEX IF NOT EXISTS idx_links_expires_at ON links(expires_at) WHERE is_active = true")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_links_shortened ON links(shortened) WHERE is_active = true")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_links_created_by ON links(created_by)")

	// Secret keys index
	db.Exec("CREATE INDEX IF NOT EXISTS idx_secret_keys_key ON secret_keys(key) WHERE is_active = true")

	// Requests indexes
	db.Exec("CREATE INDEX IF NOT EXISTS idx_requests_ip_address_requested_at ON requests(ip_address, requested_at DESC)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_requests_requested_at ON requests(requested_at DESC)")

	// Logs indexes (GORM will automatically create indexes for timestamp, type, and source due to the index tags)
	db.Exec("CREATE INDEX IF NOT EXISTS idx_logs_type_timestamp ON logs(type, timestamp DESC)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_logs_source_timestamp ON logs(source, timestamp DESC)")

	return nil
}
