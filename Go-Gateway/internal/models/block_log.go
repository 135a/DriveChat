package models

import (
	"time"

	"gorm.io/gorm"
)

// BlockLog records requests that were blocked by the gateway.
type BlockLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ClientIP  string    `gorm:"type:varchar(64);index" json:"client_ip"`
	ReqPath   string    `gorm:"type:varchar(255);index" json:"req_path"`
	RuleType  string    `gorm:"type:varchar(32)" json:"rule_type"`   // "IP_BLACKLIST", "PATH_BLACKLIST", "RATE_LIMIT_AUTO_BLOCK"
	RuleValue string    `gorm:"type:varchar(255)" json:"rule_value"` // The matched rule value
	CreatedAt time.Time `gorm:"index" json:"created_at"`             // Indexed for time-based filtering in dashboard
}

// Disable soft-delete for logs if we want to save space, or just use regular Delete.
// Usually logs don't need DeletedAt.
func (BlockLog) BeforeCreate(tx *gorm.DB) (err error) {
	// Custom hooks if needed
	return
}
