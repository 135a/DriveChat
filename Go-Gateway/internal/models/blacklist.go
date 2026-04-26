package models

import (
	"time"

	"gorm.io/gorm"
)

// BlacklistRule defines IPs or Paths that are blocked.
type BlacklistRule struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Type      string         `gorm:"type:varchar(32);index;not null" json:"type"` // e.g., "IP", "PATH_PREFIX", "PATH_EXACT"
	Value     string         `gorm:"type:varchar(255);index;not null" json:"value"` // The actual IP or path
	IsAuto    bool           `gorm:"default:false" json:"is_auto"`                // True if added by auto-blacklist mechanism
	Status    int            `gorm:"type:tinyint;default:1" json:"status"`        // 1: active, 0: inactive
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
