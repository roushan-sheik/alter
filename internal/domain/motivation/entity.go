package motivation

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Motivation represents the MOTIVATIONS table.
type Motivation struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey"`
	Title        string    `gorm:"type:varchar(255);not null"`
	SpeakerName  string    `gorm:"type:varchar(255);not null"`
	Description  string    `gorm:"type:text;not null"`
	VideoURL     string    `gorm:"type:text;not null"`
	ThumbnailURL string    `gorm:"type:text;not null"`
	Duration     string    `gorm:"type:varchar(20);not null"` // e.g., "18:24"
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

// BeforeCreate assigns a UUID before inserting a new Motivation row.
func (m *Motivation) BeforeCreate(_ *gorm.DB) error {
	m.ID = uuid.New()
	return nil
}
