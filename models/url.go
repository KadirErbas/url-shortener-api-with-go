package models

import (
	"time"
)

type URL struct {
	ID          uint           	`gorm:"primaryKey"`
	OriginalURL string         	`gorm:"not null;unique"`
	ShortURL    string         	`gorm:"not null;unique"`
	CreatedAt   time.Time      	`json:"created_at"`
	ExpiresAt   time.Time      	`json:"expires_at"`
	DeletedAt  	*time.Time		`json:"deleted_at"`
	UsageCount  int            	`json:"usage_count"`
}
