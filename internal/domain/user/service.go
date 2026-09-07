package user

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"log"
	"math/big"
	"time"

	"gotickets/internal/auth"
	"gotickets/internal/domain/user/dto"
	"gotickets/internal/email"
	"gotickets/internal/upload"

	firebaseAuth "firebase.google.com/go/v4/auth"
	"github.com/google/uuid"
)

var (
	ErrDuplicateEmail      = errors.New("email is already registered")
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrInvalidRefreshToken = errors.New("invalid or expired refresh token")
	ErrInvalidOTP          = errors.New("OTP is invalid, expired, or already used")
	ErrAccountDeleted      = errors.New("account has been deleted")
	ErrRateLimited         = errors.New("too many requests — please try again later")
)

// Service defines the business logic contract for the user domain.
type Service interface {
	// Register creates a new email-provider account and returns the full standard response.
	Register(req dto.RegisterRequest) (*dto.StandardAuthResponse, error)
	// Login authenticates an email user and returns the full standard response.
	Login(req dto.LoginRequest) (*dto.StandardAuthResponse, error)
	SocialLogin(ctx context.Context, req dto.SocialLoginRequest) (*dto.AuthResponse, error)
	RefreshToken(rawToken string) (*dto.AuthResponse, error)
	Logout(rawToken string) error
	ForgotPassword(req dto.ForgotPasswordRequest) error
	ResendOTP(req dto.ResendOTPRequest) error
	VerifyOTP(req dto.VerifyOTPRequest) (*dto.VerifyOTPResponse, error)
	ResetPassword(req dto.ResetPasswordRequest) error

	GetProfileByEmail(email string) (*dto.ProfileResponse, error)
	UpdateProfileByEmail(email string, req dto.UpdateProfileRequest) (*dto.ProfileResponse, error)
	UpdateAvatarURLByEmail(email string, avatarURL string) (*dto.ProfileResponse, error)
	ChangePasswordByEmail(email string, req dto.ChangePasswordRequest) error
	DeleteAccountByEmail(email string) error
	GetUserIDByEmail(email string) (uuid.UUID, error)

	RegisterDevice(userID uuid.UUID, req dto.RegisterDeviceRequest) error

	AdminLogin(email, password string) (*dto.AuthResponse, error)
}

type service struct {
	repo          Repository
	jwt           auth.JWTService
	uploader      upload.Uploader
	mailer        email.Mailer
	firebaseAuth  *firebaseAuth.Client
	limiter       *RateLimiter
	adminEmail    string
	adminPassword string
}

// NewService creates a new user Service. uploader, mailer, and firebaseAuth may be nil.
func NewService(repo Repository, jwt auth.JWTService, uploader upload.Uploader, mailer email.Mailer, fbAuth *firebaseAuth.Client, adminEmail, adminPassword string) Service {
	return &service{repo: repo, jwt: jwt, uploader: uploader, mailer: mailer, firebaseAuth: fbAuth, limiter: newRateLimiter(), adminEmail: adminEmail, adminPassword: adminPassword}
}


func (s *service) Register(req dto.RegisterRequest) (*dto.StandardAuthResponse, error) {
	// Check for duplicate email
	existing, err := s.repo.GetUserByEmail(req.Email)
	if err == nil && existing != nil {
		return nil, ErrDuplicateEmail
	}

	var termsAcceptedAt *time.Time
	if req.AgreeTermsAndConditions {
		now := time.Now()
		termsAcceptedAt = &now
	}

	u := &User{
		Name:               req.Name,
		Email:              req.Email,
		LanguagePreference: req.LanguagePreference,
		Age:                req.Age,
		TermsAcceptedAt:    termsAcceptedAt,
		AuthProvider:       AuthProviderEmail,
	}
	if err := u.HashPassword(req.Password); err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	if err := s.repo.CreateUser(u); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return s.issueStandardResponse(u, "Registration successful")
}

func (s *service) Login(req dto.LoginRequest) (*dto.StandardAuthResponse, error) {
	if !s.limiter.AllowLogin(req.Email) {
		return nil, ErrRateLimited
	}
	u, err := s.repo.GetUserByEmail(req.Email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	if err := u.CheckPassword(req.Password); err != nil {
		return nil, ErrInvalidCredentials
	}

	if req.LanguagePreference != "" && req.LanguagePreference != u.LanguagePreference {
		u.LanguagePreference = req.LanguagePreference
		if err := s.repo.UpdateUser(u); err != nil {
			return nil, fmt.Errorf("failed to update language preference: %w", err)
		}
	}

	return s.issueStandardResponse(u, "Login successful")
}

func (s *service) SocialLogin(ctx context.Context, req dto.SocialLoginRequest) (*dto.AuthResponse, error) {
	if s.firebaseAuth == nil {
		return nil, errors.New("social login is not configured")
	}

	token, err := s.firebaseAuth.VerifyIDToken(ctx, req.IDToken)
	if err != nil {
		return nil, fmt.Errorf("invalid id_token: %w", err)
	}

	email, ok := token.Claims["email"].(string)
	if !ok || email == "" {
		return nil, errors.New("email not found in id_token")
	}

	u, err := s.repo.GetUserByEmail(email)
	if err != nil {
		// User does not exist, create a new one
		provider := AuthProviderGoogle
		if req.Provider == "apple" {
			provider = AuthProviderApple
		}

		u = &User{
			Name:         req.Name,
			Email:        email,
			AuthProvider: provider,
			// PasswordHash left as nil
		}

		if err := s.repo.CreateUser(u); err != nil {
			return nil, fmt.Errorf("failed to create social user: %w", err)
		}
	}

	return s.issueTokenPair(u)
}

func (s *service) AdminLogin(email, password string) (*dto.AuthResponse, error) {
	if !s.limiter.AllowLogin(email) {
		return nil, ErrRateLimited
	}

	u, err := s.repo.GetUserByEmail(email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if err := u.CheckPassword(password); err != nil {
		return nil, ErrInvalidCredentials
	}

	if u.Role != RoleAdmin {
		return nil, errors.New("unauthorized: account is not an admin")
	}

	return s.issueTokenPair(u)
}

func (s *service) RefreshToken(rawToken string) (*dto.AuthResponse, error) {

	_, err := s.jwt.ValidateToken(rawToken, true)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	hash := hashToken(rawToken)
	rt, err := s.repo.GetRefreshToken(hash)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	if err := s.repo.RevokeRefreshToken(rt.TokenHash); err != nil {
		return nil, fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	u, err := s.repo.GetUserByID(rt.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return s.issueTokenPair(u)
}

func (s *service) Logout(rawToken string) error {
	if rawToken == "" {
		return nil
	}
	hash := hashToken(rawToken)
	return s.repo.RevokeRefreshToken(hash)
}

func (s *service) ForgotPassword(req dto.ForgotPasswordRequest) error {
	if !s.limiter.AllowForgotPassword(req.Email) {
		return ErrRateLimited
	}

	u, err := s.repo.GetUserByEmail(req.Email)
	if err != nil {
		return nil
	}

	_ = s.repo.InvalidatePendingOTPs(req.Email)

	code, err := generateOTPCode()
	if err != nil {
		return fmt.Errorf("failed to generate OTP: %w", err)
	}

	otp := &OTP{
		Email:     req.Email,
		Code:      code,
		Purpose:   "PASSWORD_RESET",
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}
	if u != nil {
		otp.UserID = &u.ID
	}

	if err := s.repo.CreateOTP(otp); err != nil {
		return fmt.Errorf("failed to create OTP: %w", err)
	}

	// Send OTP via Gmail SMTP when mailer is configured; log to stdout as fallback.
	if s.mailer != nil {
		if err := s.mailer.SendOTP(u.Email, u.Name, code); err != nil {
			// Non-fatal: OTP is already stored. Log and continue — the user can retry.
			log.Printf("[EMAIL] failed to send OTP to %s: %v", u.Email, err)
		}
	} else {
		// [DEV_MODE] No mailer configured — print OTP to stdout.
		fmt.Printf("[OTP][DEV_MODE] Email: %s Code: %s (expires in 10m)\n", req.Email, code)
	}

	return nil
}

func (s *service) ResendOTP(req dto.ResendOTPRequest) error {
	if !s.limiter.AllowResendOTP(req.Email) {
		return ErrRateLimited
	}

	u, err := s.repo.GetUserByEmail(req.Email)
	if err != nil {
		return nil
	}

	_ = s.repo.InvalidatePendingOTPs(req.Email)

	code, err := generateOTPCode()
	if err != nil {
		return fmt.Errorf("failed to generate OTP: %w", err)
	}

	otp := &OTP{
		Email:     req.Email,
		Code:      code,
		Purpose:   "PASSWORD_RESET",
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}
	if u != nil {
		otp.UserID = &u.ID
	}

	if err := s.repo.CreateOTP(otp); err != nil {
		return fmt.Errorf("failed to create OTP: %w", err)
	}

	if s.mailer != nil {
		if err := s.mailer.SendOTP(u.Email, u.Name, code); err != nil {
			log.Printf("[EMAIL] failed to send OTP to %s: %v", u.Email, err)
		}
	} else {
		fmt.Printf("[OTP][DEV_MODE] Email: %s Code: %s (expires in 10m)\n", req.Email, code)
	}

	return nil
}

func (s *service) VerifyOTP(req dto.VerifyOTPRequest) (*dto.VerifyOTPResponse, error) {
	otp, err := s.repo.GetValidOTP(req.Email, req.OTP)
	if err != nil {
		return nil, ErrInvalidOTP
	}

	u, err := s.repo.GetUserByEmail(req.Email)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if err := s.repo.MarkOTPUsed(otp.ID); err != nil {
		return nil, fmt.Errorf("failed to mark OTP used: %w", err)
	}

	resetToken, err := s.jwt.GenerateResetToken(u.ID, u.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate reset token: %w", err)
	}

	return &dto.VerifyOTPResponse{
		ResetToken: resetToken,
	}, nil
}

func (s *service) ResetPassword(req dto.ResetPasswordRequest) error {
	// TODO: add rate limiting here (e.g. Redis-backed per-limiter if necessary)

	claims, err := s.jwt.ValidateResetToken(req.ResetToken)
	if err != nil {
		return errors.New("invalid or expired reset token")
	}

	u, err := s.repo.GetUserByID(claims.UserID)
	if err != nil {
		return errors.New("user not found")
	}

	if err := u.HashPassword(req.NewPassword); err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	if err := s.repo.UpdateUser(u); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Invalidate existing sessions
	if err := s.repo.RevokeAllRefreshTokens(u.ID); err != nil {
		log.Printf("[WARNING] failed to revoke refresh tokens for user %s: %v", u.ID, err)
	}

	return nil
}


func (s *service) GetProfileByEmail(email string) (*dto.ProfileResponse, error) {
	u, err := s.repo.GetUserByEmail(email)
	if err != nil {
		return nil, err
	}
	return toProfileResponse(u), nil
}

func (s *service) GetUserIDByEmail(email string) (uuid.UUID, error) {
	u, err := s.repo.GetUserByEmail(email)
	if err != nil {
		return uuid.Nil, err
	}
	return u.ID, nil
}

func (s *service) UpdateProfileByEmail(email string, req dto.UpdateProfileRequest) (*dto.ProfileResponse, error) {
	u, err := s.repo.GetUserByEmail(email)
	if err != nil {
		return nil, err
	}
	if req.Name != nil {
		u.Name = *req.Name
	}
	if req.Location != nil {
		u.Location = req.Location
	}
	if req.ThemePreference != nil {
		u.ThemePreference = ThemePreference(*req.ThemePreference)
	}
	if req.LanguagePreference != nil {
		u.LanguagePreference = *req.LanguagePreference
	}
	if req.Age != nil {
		u.Age = *req.Age
	}
	if err := s.repo.UpdateUser(u); err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}
	return toProfileResponse(u), nil
}

func (s *service) UpdateAvatarURLByEmail(email string, avatarURL string) (*dto.ProfileResponse, error) {
	u, err := s.repo.GetUserByEmail(email)
	if err != nil {
		return nil, err
	}
	u.AvatarURL = &avatarURL
	if err := s.repo.UpdateUser(u); err != nil {
		return nil, fmt.Errorf("failed to update avatar: %w", err)
	}
	return toProfileResponse(u), nil
}

func (s *service) ChangePasswordByEmail(email string, req dto.ChangePasswordRequest) error {
	u, err := s.repo.GetUserByEmail(email)
	if err != nil {
		return err
	}
	if err := u.CheckPassword(req.OldPassword); err != nil {
		return errors.New("current password is incorrect")
	}
	if err := u.HashPassword(req.NewPassword); err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	return s.repo.UpdateUser(u)
}

func (s *service) DeleteAccountByEmail(email string) error {
	u, err := s.repo.GetUserByEmail(email)
	if err != nil {
		return err
	}
	return s.repo.SoftDeleteUser(u.ID)
}

func (s *service) ChangePassword(userID uuid.UUID, req dto.ChangePasswordRequest) error {
	u, err := s.repo.GetUserByID(userID)
	if err != nil {
		return err
	}
	if err := u.CheckPassword(req.OldPassword); err != nil {
		return errors.New("current password is incorrect")
	}
	if err := u.HashPassword(req.NewPassword); err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	return s.repo.UpdateUser(u)
}

func (s *service) DeleteAccount(userID uuid.UUID) error {
	return s.repo.SoftDeleteUser(userID)
}

func (s *service) UpdateAvatar(ctx context.Context, userID uuid.UUID, file interface{}, _ interface{}) (*dto.AvatarResponse, error) {
	if s.uploader == nil {
		return nil, errors.New("upload service not configured")
	}
	return nil, errors.New("use handler-level UploadAvatar — this method is a handler-level concern")
}


func (s *service) RegisterDevice(userID uuid.UUID, req dto.RegisterDeviceRequest) error {
	dt := &DeviceToken{
		UserID:   userID,
		Token:    req.Token,
		Platform: Platform(req.Platform),
	}
	return s.repo.UpsertDeviceToken(dt)
}


// issueStandardResponse generates tokens and builds the full StandardAuthResponse
// envelope used by the Login and Register endpoints.
func (s *service) issueStandardResponse(u *User, message string) (*dto.StandardAuthResponse, error) {
	accessToken, refreshToken, err := s.jwt.GenerateToken(u.ID, u.Name, u.Email, string(u.Role), u.IsPremium)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	rt := &RefreshToken{
		UserID:    u.ID,
		TokenHash: hashToken(refreshToken),
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
	}
	if err := s.repo.CreateRefreshToken(rt); err != nil {
		return nil, fmt.Errorf("failed to persist refresh token: %w", err)
	}

	// Resolve avatar URL (nil pointer → empty string for the DTO).
	avatarURL := ""
	if u.AvatarURL != nil {
		avatarURL = *u.AvatarURL
	}

	return &dto.StandardAuthResponse{
		Success: true,
		Message: message,
		Data: dto.AuthDataResponse{
			User: dto.UserDTO{
				ID:        u.ID,
				Name:      u.Name,
				Email:     u.Email,
				AvatarURL: avatarURL,
				Role:      string(u.Role),
			},
			Tokens: dto.TokenDTO{
				AccessToken:  accessToken,
				RefreshToken: refreshToken,
			},
		},
	}, nil
}

// issueTokenPair generates a new JWT access+refresh token pair, persists the refresh token, and returns both.
// Used by Refresh, AdminLogin, and SocialLogin which return the lightweight AuthResponse.
func (s *service) issueTokenPair(u *User) (*dto.AuthResponse, error) {
	accessToken, refreshToken, err := s.jwt.GenerateToken(u.ID, u.Name, u.Email, string(u.Role), u.IsPremium)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	rt := &RefreshToken{
		UserID:    u.ID,
		TokenHash: hashToken(refreshToken),
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
	}
	if err := s.repo.CreateRefreshToken(rt); err != nil {
		return nil, fmt.Errorf("failed to persist refresh token: %w", err)
	}

	return &dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// hashToken returns a SHA-256 hex hash of the token string for safe DB storage.
func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", h)
}

// generateOTPCode generates a random 5-digit numeric string.
func generateOTPCode() (string, error) {
	const digits = "0123456789"
	code := make([]byte, 5)
	for i := range code {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", err
		}
		code[i] = digits[n.Int64()]
	}
	return string(code), nil
}

// toProfileResponse maps a User entity to a ProfileResponse DTO.
func toProfileResponse(u *User) *dto.ProfileResponse {
	return &dto.ProfileResponse{
		ID:                 u.ID,
		Name:               u.Name,
		Email:              u.Email,
		AuthProvider:       string(u.AuthProvider),
		Location:           u.Location,
		AvatarURL:          u.AvatarURL,
		ThemePreference:    string(u.ThemePreference),
		LanguagePreference: u.LanguagePreference,
		Age:                u.Age,
		IsPremium:          u.IsPremium,
		TermsAcceptedAt:    u.TermsAcceptedAt,
		CreatedAt:          u.CreatedAt,
	}
}
