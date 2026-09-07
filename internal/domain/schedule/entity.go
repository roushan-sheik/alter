package schedule

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserSchedule represents the USER_SCHEDULES table.
// One-to-one relationship with users (enforced by unique index on user_id).
type UserSchedule struct {
	ID                uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID            uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"` // 1:1 with users
	MorningPrayerTime string    `gorm:"type:varchar(8);not null;index"` // "HH:MM:SS"
	NightPrayerTime   string    `gorm:"type:varchar(8);not null;index"` // "HH:MM:SS"
	Timezone          string    `gorm:"type:varchar(100);not null"`     // IANA tz, e.g. "Asia/Dhaka"
	PushEnabled       bool      `gorm:"not null;default:true"`
	UpdatedAt         time.Time
}

// BeforeCreate assigns a UUID before inserting a new UserSchedule row.
func (s *UserSchedule) BeforeCreate(_ *gorm.DB) error {
	s.ID = uuid.New()
	return nil
}
