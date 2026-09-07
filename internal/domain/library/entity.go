package library

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LibraryItem struct {
	ID               uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Title            string         `gorm:"type:varchar(255);not null"`
	Category         string         `gorm:"type:varchar(100);not null"`
	ShortDescription string         `gorm:"type:text;not null"`
	ContentText      string         `gorm:"type:text;not null"`
	ThumbnailURL     string         `gorm:"type:varchar(255);not null"`
	MediaURL         *string        `gorm:"type:varchar(255)"` // Nullable for optional video/audio
	CreatedAt        time.Time      `gorm:"autoCreateTime"`
	UpdatedAt        time.Time      `gorm:"autoUpdateTime"`
	DeletedAt        gorm.DeletedAt `gorm:"index"`
}

type LibraryCategory struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name      string         `gorm:"type:varchar(100);not null;uniqueIndex"`
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
