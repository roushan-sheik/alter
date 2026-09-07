package user

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AuthProvider represents the login provider.
type AuthProvider string

// ThemePreference represents the user's chosen UI theme.
type ThemePreference string

// Platform represents the device platform for push notifications.
type Platform string

// Role represents the user's role.
type Role string

const (
	AuthProviderEmail  AuthProvider = "EMAIL"
	AuthProviderGoogle AuthProvider = "GOOGLE"
	AuthProviderApple  AuthProvider = "APPLE"

	RoleUser  Role = "USER"
	RoleAdmin Role = "ADMIN"

	ThemeLight ThemePreference = "LIGHT"
	ThemeDark  ThemePreference = "DARK"

	PlatformIOS     Platform = "IOS"
	PlatformAndroid Platform = "ANDROID"
)

// User represents the USERS table.
type User struct {
	ID                 uuid.UUID       `gorm:"type:uuid;primaryKey"`
	Name               string          `gorm:"type:varchar(255);not null"`
	Email              string          `gorm:"type:varchar(255);uniqueIndex;not null"`
	PasswordHash       *string         `gorm:"type:varchar(255)"` // Nullable for OAuth users
	AuthProvider       AuthProvider    `gorm:"type:varchar(20);not null;default:'EMAIL'"`
	Role               Role            `gorm:"type:varchar(20);not null;default:'USER'"`
	Location           *string         `gorm:"type:varchar(255)"`
	AvatarURL          *string         `gorm:"type:text"`
	ThemePreference    ThemePreference `gorm:"type:varchar(20);not null;default:'LIGHT'"`
	LanguagePreference string          `gorm:"type:varchar(10);not null;default:'en'"`
	Age                int             `gorm:"not null;default:0"`
	IsPremium          bool            `gorm:"not null;default:false"`
	TermsAcceptedAt    *time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
	DeletedAt          gorm.DeletedAt `gorm:"index"` // Soft-delete for GDPR grace period
}

// BeforeCreate assigns a UUID before inserting a new User row.
func (u *User) BeforeCreate(_ *gorm.DB) error {
	u.ID = uuid.New()
	return nil
}

// HashPassword hashes the plain text password using bcrypt cost 12.
func (u *User) HashPassword(plain string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), 12)
	if err != nil {
		return err
	}
	h := string(hash)
	u.PasswordHash = &h
	return nil
}

// CheckPassword compares a plain text password against the stored hash.
func (u *User) CheckPassword(plain string) error {
	if u.PasswordHash == nil {
		return bcrypt.ErrMismatchedHashAndPassword
	}
	return bcrypt.CompareHashAndPassword([]byte(*u.PasswordHash), []byte(plain))
}

// RefreshToken represents the REFRESH_TOKENS table.
type RefreshToken struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null;index"`
	TokenHash string     `gorm:"type:varchar(255);not null;uniqueIndex"`
	ExpiresAt time.Time  `gorm:"not null"`
	RevokedAt *time.Time // Nullable
	CreatedAt time.Time
}

// BeforeCreate assigns a UUID before inserting a new RefreshToken row.
func (rt *RefreshToken) BeforeCreate(_ *gorm.DB) error {
	rt.ID = uuid.New()
	return nil
}

// OTP represents the OTPS table.
type OTP struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey"`
	UserID    *uuid.UUID `gorm:"type:uuid;index"` // Nullable until verified against an existing user
	Email     string     `gorm:"type:varchar(255);not null;index"`
	Code      string     `gorm:"type:varchar(5);not null"` // 5-digit
	Purpose   string     `gorm:"type:varchar(50);not null;default:'PASSWORD_RESET'"`
	ExpiresAt time.Time  `gorm:"not null"`
	IsUsed    bool       `gorm:"not null;default:false"`
	CreatedAt time.Time
}

// BeforeCreate assigns a UUID before inserting a new OTP row.
func (o *OTP) BeforeCreate(_ *gorm.DB) error {
	o.ID = uuid.New()
	return nil
}

// DeviceToken represents the DEVICE_TOKENS table.
type DeviceToken struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID     uuid.UUID `gorm:"type:uuid;not null;index"`
	Token      string    `gorm:"type:varchar(500);not null;uniqueIndex"`
	Platform   Platform  `gorm:"type:varchar(20);not null"`
	CreatedAt  time.Time
	LastSeenAt time.Time
}

// BeforeCreate assigns a UUID before inserting a new DeviceToken row.
func (dt *DeviceToken) BeforeCreate(_ *gorm.DB) error {
	dt.ID = uuid.New()
	return nil
}
