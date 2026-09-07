package user

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	// User
	CreateUser(u *User) error
	GetUserByEmail(email string) (*User, error)
	GetUserByID(id uuid.UUID) (*User, error)
	UpdateUser(u *User) error
	SoftDeleteUser(id uuid.UUID) error

	// Refresh tokens
	CreateRefreshToken(rt *RefreshToken) error
	GetRefreshToken(tokenHash string) (*RefreshToken, error)
	RevokeRefreshToken(tokenHash string) error
	RevokeAllRefreshTokens(userID uuid.UUID) error

	// OTPs
	InvalidatePendingOTPs(email string) error
	CreateOTP(otp *OTP) error
	GetValidOTP(email, code string) (*OTP, error)
	MarkOTPUsed(id uuid.UUID) error

	// Devices
	UpsertDeviceToken(dt *DeviceToken) error
}


type repository struct {
	db *gorm.DB
}

// NewRepository returns a GORM-backed Repository.
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateUser(u *User) error {
	return r.db.Create(u).Error
}

func (r *repository) GetUserByEmail(email string) (*User, error) {
	var u User
	err := r.db.Where("email = ?", email).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *repository) GetUserByID(id uuid.UUID) (*User, error) {
	var u User
	err := r.db.Where("id = ?", id).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *repository) UpdateUser(u *User) error {
	return r.db.Save(u).Error
}

func (r *repository) SoftDeleteUser(id uuid.UUID) error {
	return r.db.Delete(&User{}, "id = ?", id).Error
}


func (r *repository) CreateRefreshToken(rt *RefreshToken) error {
	return r.db.Create(rt).Error
}

func (r *repository) GetRefreshToken(tokenHash string) (*RefreshToken, error) {
	var rt RefreshToken
	err := r.db.Where("token_hash = ? AND revoked_at IS NULL AND expires_at > ?", tokenHash, time.Now()).First(&rt).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("refresh token not found or revoked")
		}
		return nil, err
	}
	return &rt, nil
}

func (r *repository) RevokeRefreshToken(tokenHash string) error {
	now := time.Now()
	return r.db.Model(&RefreshToken{}).
		Where("token_hash = ?", tokenHash).
		Update("revoked_at", now).Error
}

func (r *repository) RevokeAllRefreshTokens(userID uuid.UUID) error {
	now := time.Now()
	return r.db.Model(&RefreshToken{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", now).Error
}


func (r *repository) InvalidatePendingOTPs(email string) error {
	now := time.Now()
	return r.db.Model(&OTP{}).
		Where("email = ? AND is_used = false AND expires_at > ?", email, now).
		Update("is_used", true).Error
}

func (r *repository) CreateOTP(otp *OTP) error {
	return r.db.Create(otp).Error
}

func (r *repository) GetValidOTP(email, code string) (*OTP, error) {
	var otp OTP
	err := r.db.Where("email = ? AND code = ? AND is_used = false AND expires_at > ?", email, code, time.Now()).
		First(&otp).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("OTP is invalid, expired, or already used")
		}
		return nil, err
	}
	return &otp, nil
}

func (r *repository) MarkOTPUsed(id uuid.UUID) error {
	return r.db.Model(&OTP{}).Where("id = ?", id).Update("is_used", true).Error
}


func (r *repository) UpsertDeviceToken(dt *DeviceToken) error {
	var existing DeviceToken
	err := r.db.Where("user_id = ? AND token = ?", dt.UserID, dt.Token).First(&existing).Error
	if err == nil {
		// Token already registered — update last_seen_at and platform
		existing.Platform = dt.Platform
		existing.LastSeenAt = time.Now()
		return r.db.Save(&existing).Error
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		dt.LastSeenAt = time.Now()
		return r.db.Create(dt).Error
	}
	return err
}
