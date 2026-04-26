package models

import (
	"time"

	"gorm.io/gorm"
)

// RouteRule defines a dynamic routing rule for the gateway.
type RouteRule struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	PathPrefix string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"path_prefix"` // e.g., /api/v1
	TargetURL  string         `gorm:"type:varchar(255);not null" json:"target_url"`             // e.g., http://backend-service:8080
	Status     int            `gorm:"type:tinyint;default:1" json:"status"`                     // 1: active, 0: inactive
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}
