package user

import (
	"errors"
	"net/http"
	"time"

	"gotickets/internal/auth"
	"gotickets/internal/domain/user/dto"
	"gotickets/internal/httpresponse"
	"gotickets/internal/upload"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

// maxAvatarSize limits avatar uploads to 5 MB.
// NOTE: Content media uploads will need a larger limit — adjust at that point.
const maxAvatarSize = 5 << 20

// Handler holds the Echo HTTP handlers for the user domain.
type Handler struct {
	svc      Service
	uploader upload.Uploader
}

// NewHandler creates a new user Handler.
func NewHandler(svc Service, uploader upload.Uploader) *Handler {
	return &Handler{svc: svc, uploader: uploader}
}


// Register godoc
// @Summary      Register a new user
// @Description  Creates a new EMAIL-provider account. Returns user profile + access/refresh JWT pair. Duplicate email returns 409. Available languages: en (English), fr (French), es (Spanish), pt (Portuguese), ht (Haitian Creole). Available auth providers: EMAIL, GOOGLE, APPLE.
// @Tags         1. Auth - Onboarding
// @Accept       json
// @Produce      json
// @Param        request  body      dto.RegisterRequest        true  "Registration payload"
// @Success      201      {object}  dto.StandardAuthResponse        "Registration successful"
// @Failure      400      {object}  httpresponse.Error  "Validation error"
// @Failure      409      {object}  httpresponse.Error  "Email already registered"
// @Failure      500      {object}  httpresponse.Error
// @Router       /api/v1/auth/register [post]
func (h *Handler) Register(c *echo.Context) error {
	var req dto.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, httpresponse.NewError(http.StatusBadRequest, "Invalid request body", err.Error()))
	}
	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, httpresponse.NewError(http.StatusBadRequest, "Validation failed", err.Error()))
	}

	resp, err := h.svc.Register(req)
	if err != nil {
		if errors.Is(err, ErrDuplicateEmail) {
			return c.JSON(http.StatusConflict, httpresponse.NewError(http.StatusConflict, "Email already registered", ""))
		}
		return c.JSON(http.StatusInternalServerError, httpresponse.NewError(http.StatusInternalServerError, "Registration failed", err.Error()))
	}

	setTokenCookies(c, resp.Data.Tokens.AccessToken, resp.Data.Tokens.RefreshToken)
	return c.JSON(http.StatusCreated, resp)
}

// Login godoc
// @Summary      Login
// @Description  Authenticates an EMAIL user. Returns user profile + access/refresh JWT pair. Available languages: en (English), fr (French), es (Spanish), pt (Portuguese), ht (Haitian Creole).
// @Tags         1. Auth - Onboarding
// @Accept       json
// @Produce      json
// @Param        request  body      dto.LoginRequest          true  "Login payload"
// @Success      200      {object}  dto.StandardAuthResponse       "Login successful"
// @Failure      400      {object}  httpresponse.Error  "Validation error"
// @Failure      401      {object}  httpresponse.Error  "Invalid credentials"
// @Router       /api/v1/auth/login [post]
func (h *Handler) Login(c *echo.Context) error {
	var req dto.LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, httpresponse.NewError(http.StatusBadRequest, "Invalid request body", err.Error()))
	}
	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, httpresponse.NewError(http.StatusBadRequest, "Validation failed", err.Error()))
	}

	resp, err := h.svc.Login(req)
	if err != nil {
		if errors.Is(err, ErrRateLimited) {
			return c.JSON(http.StatusTooManyRequests, httpresponse.NewError(http.StatusTooManyRequests, "Too many login attempts", "Please try again later"))
		}
		if errors.Is(err, ErrInvalidCredentials) {
			return c.JSON(http.StatusUnauthorized, httpresponse.NewError(http.StatusUnauthorized, "Invalid email or password", ""))
		}
		return c.JSON(http.StatusInternalServerError, httpresponse.NewError(http.StatusInternalServerError, "Login failed", err.Error()))
	}

	setTokenCookies(c, resp.Data.Tokens.AccessToken, resp.Data.Tokens.RefreshToken)
	return c.JSON(http.StatusOK, resp)
}

// AdminLogin godoc
// @Summary      Admin Login
// @Description  Authenticates an admin using hardcoded credentials. Returns access + refresh JWT pair.
// @Tags         1. Auth - Onboarding
// @Accept       json
// @Produce      json
// @Param        request  body      dto.AdminLoginRequest  true  "Admin Login payload"
// @Success      200      {object}  dto.AuthResponse
// @Failure      400      {object}  httpresponse.Error  "Validation error"
// @Failure      401      {object}  httpresponse.Error  "Invalid credentials"
// @Router       /api/v1/auth/admin/login [post]
func (h *Handler) AdminLogin(c *echo.Context) error {
	var req dto.AdminLoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, httpresponse.NewError(http.StatusBadRequest, "Invalid request body", err.Error()))
	}
	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, httpresponse.NewError(http.StatusBadRequest, "Validation failed", err.Error()))
	}

	resp, err := h.svc.AdminLogin(req.Email, req.Password)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, httpresponse.NewError(http.StatusUnauthorized, "Invalid admin email or password", ""))
	}

	setTokenCookies(c, resp.AccessToken, resp.RefreshToken)
	return c.JSON(http.StatusOK, resp)
}

// SocialLogin godoc
// @Summary      Social Login
// @Description  Authenticates a user using Firebase ID token (Google or Apple). Returns access + refresh JWT pair.
// @Tags         1. Auth - Onboarding
// @Accept       json
// @Produce      json
// @Param        request  body      dto.SocialLoginRequest  true  "Social Login payload"
// @Success      200      {object}  dto.AuthResponse
// @Failure      400      {object}  httpresponse.Error  "Validation error"
// @Failure      401      {object}  httpresponse.Error  "Invalid token"
// @Router       /api/v1/auth/social-login [post]
func (h *Handler) SocialLogin(c *echo.Context) error {
	var req dto.SocialLoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, httpresponse.NewError(http.StatusBadRequest, "Invalid request body", err.Error()))
	}
	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, httpresponse.NewError(http.StatusBadRequest, "Validation failed", err.Error()))
	}

	resp, err := h.svc.SocialLogin(c.Request().Context(), req)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, httpresponse.NewError(http.StatusUnauthorized, "Social login failed", err.Error()))
	}

	setTokenCookies(c, resp.AccessToken, resp.RefreshToken)
	return c.JSON(http.StatusOK, resp)
}

// Refresh godoc
// @Summary      Refresh access token
// @Description  Rotates the refresh token: revokes the old one and issues a new access + refresh pair. Token may be sent in body or "refresh_token" cookie.
// @Tags         3. Auth - Session Management
// @Accept       json
// @Produce      json
// @Param        request  body      dto.RefreshRequest  false  "Refresh token (omit if using cookie)"
// @Success      200      {object}  dto.AuthResponse
// @Failure      401      {object}  httpresponse.Error  "Invalid or expired refresh token"
// @Router       /api/v1/auth/refresh [post]
func (h *Handler) Refresh(c *echo.Context) error {
	token := readRefreshToken(c)
	if token == "" {
		return c.JSON(http.StatusUnauthorized, httpresponse.NewError(http.StatusUnauthorized, "Refresh token required", ""))
	}

	resp, err := h.svc.RefreshToken(token)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, httpresponse.NewError(http.StatusUnauthorized, "Invalid or expired refresh token", ""))
	}

	setTokenCookies(c, resp.AccessToken, resp.RefreshToken)
	return c.JSON(http.StatusOK, resp)
}

// Logout godoc
// @Summary      Logout
// @Description  Revokes the current refresh token, ending the session.
// @Tags         3. Auth - Session Management
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      dto.RefreshRequest  false  "Refresh token (omit if using cookie)"
// @Success      200      {object}  dto.MessageResponse
// @Failure      401      {object}  httpresponse.Error
// @Router       /api/v1/auth/logout [post]
func (h *Handler) Logout(c *echo.Context) error {
	token := readRefreshToken(c)
	if err := h.svc.Logout(token); err != nil {
		return c.JSON(http.StatusInternalServerError, httpresponse.NewError(http.StatusInternalServerError, "Logout failed", err.Error()))
	}
	clearTokenCookies(c)
	return c.JSON(http.StatusOK, dto.MessageResponse{Message: "Logged out successfully"})
}

// ForgotPassword godoc
// @Summary      Request password reset OTP
// @Description  Sends a 5-digit OTP to the email (10-min expiry). Always returns 200 to prevent user enumeration.
// @Tags         2. Auth - Password Recovery
// @Accept       json
// @Produce      json
// @Param        request  body      dto.ForgotPasswordRequest  true  "Email address"
// @Success      200      {object}  dto.MessageResponse
// @Failure      400      {object}  httpresponse.Error  "Validation error"
// @Router       /api/v1/auth/forgot-password [post]
func (h *Handler) ForgotPassword(c *echo.Context) error {
	var req dto.ForgotPasswordRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, httpresponse.NewError(http.StatusBadRequest, "Invalid request body", err.Error()))
	}
	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, httpresponse.NewError(http.StatusBadRequest, "Validation failed", err.Error()))
	}

	// Rate-limit errors surface as 429; all other errors are swallowed (no user enumeration).
	err := h.svc.ForgotPassword(req)
	if errors.Is(err, ErrRateLimited) {
		return c.JSON(http.StatusTooManyRequests, httpresponse.NewError(http.StatusTooManyRequests, "Too many requests", "Please try again later"))
	}
	return c.JSON(http.StatusOK, dto.MessageResponse{Message: "If your email is registered, you will receive an OTP"})
}

// ResendOTP godoc
// @Summary      Resend password reset OTP
// @Description  Enforces a 1-minute cooldown, invalidates old OTPs, and sends a new 5-digit OTP to the email.
// @Tags         2. Auth - Password Recovery
// @Accept       json
// @Produce      json
// @Param        request  body      dto.ResendOTPRequest  true  "Email address"
// @Success      200      {object}  dto.MessageResponse
// @Failure      400      {object}  httpresponse.Error  "Validation error"
// @Failure      429      {object}  httpresponse.Error  "Too many requests"
// @Router       /api/v1/auth/resend-otp [post]
func (h *Handler) ResendOTP(c *echo.Context) error {
	var req dto.ResendOTPRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, httpresponse.NewError(http.StatusBadRequest, "Invalid request body", err.Error()))
	}
	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, httpresponse.NewError(http.StatusBadRequest, "Validation failed", err.Error()))
	}

	err := h.svc.ResendOTP(req)
	if errors.Is(err, ErrRateLimited) {
		return c.JSON(http.StatusTooManyRequests, httpresponse.NewError(http.StatusTooManyRequests, "Too many requests", "Please wait 1 minute before resending"))
	}
	return c.JSON(http.StatusOK, dto.MessageResponse{Message: "If your email is registered, you will receive an OTP"})
}

// VerifyOTP godoc
// @Summary      Verify OTP for password reset
// @Description  Verifies the 5-digit OTP and returns a temporary reset token.
// @Tags         2. Auth - Password Recovery
// @Accept       json
// @Produce      json
// @Param        request  body      dto.VerifyOTPRequest  true  "Email + OTP"
// @Success      200      {object}  dto.StandardResponse{data=dto.VerifyOTPResponse}
// @Failure      400      {object}  httpresponse.Error  "Invalid or expired OTP"
// @Router       /api/v1/auth/verify-otp [post]
func (h *Handler) VerifyOTP(c *echo.Context) error {
	var req dto.VerifyOTPRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, httpresponse.NewError(http.StatusBadRequest, "Invalid request body", err.Error()))
	}
	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, httpresponse.NewError(http.StatusBadRequest, "Validation failed", err.Error()))
	}

	resp, err := h.svc.VerifyOTP(req)
	if err != nil {
		if errors.Is(err, ErrInvalidOTP) {
			return c.JSON(http.StatusBadRequest, httpresponse.NewError(http.StatusBadRequest, "Invalid, expired, or already used OTP", ""))
		}
		return c.JSON(http.StatusInternalServerError, httpresponse.NewError(http.StatusInternalServerError, "Failed to verify OTP", err.Error()))
	}

	return c.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Message: "OTP verified successfully",
		Data:    resp,
	})
}

// ResetPassword godoc
// @Summary      Reset password with reset token
// @Description  Verifies the temporary reset token and updates the user password, revoking all existing sessions.
// @Tags         2. Auth - Password Recovery
// @Accept       json
// @Produce      json
// @Param        request  body      dto.ResetPasswordRequest  true  "Reset token + new password"
// @Success      200      {object}  dto.MessageResponse
// @Failure      400      {object}  httpresponse.Error  "Invalid or expired reset token"
// @Failure      500      {object}  httpresponse.Error
// @Router       /api/v1/auth/reset-password [post]
func (h *Handler) ResetPassword(c *echo.Context) error {
	var req dto.ResetPasswordRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, httpresponse.NewError(http.StatusBadRequest, "Invalid request body", err.Error()))
	}
	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, httpresponse.NewError(http.StatusBadRequest, "Validation failed", err.Error()))
	}

	if err := h.svc.ResetPassword(req); err != nil {
		if err.Error() == "invalid or expired reset token" {
			return c.JSON(http.StatusBadRequest, httpresponse.NewError(http.StatusBadRequest, "Invalid or expired reset token", ""))
		}
		return c.JSON(http.StatusInternalServerError, httpresponse.NewError(http.StatusInternalServerError, "Failed to reset password", err.Error()))
	}
	return c.JSON(http.StatusOK, dto.MessageResponse{Message: "Password reset successfully"})
}


// GetMe godoc
// @Summary      Get current user profile
// @Description  Returns the full profile and settings for the authenticated user.
// @Tags         Users
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  dto.ProfileResponse
// @Failure      401  {object}  httpresponse.Error
// @Failure      404  {object}  httpresponse.Error
// @Router       /api/v1/users/me [get]
func (h *Handler) GetMe(c *echo.Context) error {
	email, err := claimsEmail(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, httpresponse.NewError(http.StatusUnauthorized, "Unauthorized", err.Error()))
	}

	profile, err := h.svc.GetProfileByEmail(email)
	if err != nil {
		return c.JSON(http.StatusNotFound, httpresponse.NewError(http.StatusNotFound, "User not found", err.Error()))
	}
	return c.JSON(http.StatusOK, profile)
}

// UpdateMe godoc
// @Summary      Update profile
// @Description  Updates name, location, theme preference (Available: LIGHT, DARK), or language preference. Duplicate email returns 409. Available languages: en (English), fr (French), es (Spanish), pt (Portuguese), ht (Haitian Creole).
// @Tags         Users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      dto.UpdateProfileRequest  true  "Fields to update (all optional)"
// @Success      200      {object}  dto.ProfileResponse
// @Failure      400      {object}  httpresponse.Error
// @Failure      401      {object}  httpresponse.Error
// @Failure      500      {object}  httpresponse.Error
// @Router       /api/v1/users/me [put]
func (h *Handler) UpdateMe(c *echo.Context) error {
	email, err := claimsEmail(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, httpresponse.NewError(http.StatusUnauthorized, "Unauthorized", err.Error()))
	}

	var req dto.UpdateProfileRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, httpresponse.NewError(http.StatusBadRequest, "Invalid request body", err.Error()))
	}
	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, httpresponse.NewError(http.StatusBadRequest, "Validation failed", err.Error()))
	}

	profile, err := h.svc.UpdateProfileByEmail(email, req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, httpresponse.NewError(http.StatusInternalServerError, "Failed to update profile", err.Error()))
	}
	return c.JSON(http.StatusOK, profile)
}

// ChangePassword godoc
// @Summary      Change password
// @Description  Verifies the current password and updates it to the new one.
// @Tags         Users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      dto.ChangePasswordRequest  true  "Current and new passwords"
// @Success      200      {object}  dto.MessageResponse
// @Failure      400      {object}  httpresponse.Error  "Wrong current password"
// @Failure      401      {object}  httpresponse.Error
// @Failure      500      {object}  httpresponse.Error
// @Router       /api/v1/users/me/password [put]
func (h *Handler) ChangePassword(c *echo.Context) error {
	email, err := claimsEmail(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, httpresponse.NewError(http.StatusUnauthorized, "Unauthorized", err.Error()))
	}

	var req dto.ChangePasswordRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, httpresponse.NewError(http.StatusBadRequest, "Invalid request body", err.Error()))
	}
	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, httpresponse.NewError(http.StatusBadRequest, "Validation failed", err.Error()))
	}

	if err := h.svc.ChangePasswordByEmail(email, req); err != nil {
		return c.JSON(http.StatusBadRequest, httpresponse.NewError(http.StatusBadRequest, err.Error(), ""))
	}
	return c.JSON(http.StatusOK, dto.MessageResponse{Message: "Password changed successfully"})
}

// DeleteMe godoc
// @Summary      Delete account
// @Description  Soft-deletes the authenticated user's account (sets deleted_at). No hard purge occurs synchronously.
// @Tags         Users
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  dto.MessageResponse
// @Failure      401  {object}  httpresponse.Error
// @Failure      500  {object}  httpresponse.Error
// @Router       /api/v1/users/me [delete]
func (h *Handler) DeleteMe(c *echo.Context) error {
	email, err := claimsEmail(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, httpresponse.NewError(http.StatusUnauthorized, "Unauthorized", err.Error()))
	}

	if err := h.svc.DeleteAccountByEmail(email); err != nil {
		return c.JSON(http.StatusInternalServerError, httpresponse.NewError(http.StatusInternalServerError, "Failed to delete account", err.Error()))
	}
	clearTokenCookies(c)
	return c.JSON(http.StatusOK, dto.MessageResponse{Message: "Account deleted successfully"})
}

// UploadAvatar godoc
// @Summary      Upload avatar
// @Description  Uploads a new profile picture (JPEG/PNG/WebP, max 5 MB) to Cloudinary and updates avatar_url.
// @Tags         Users
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        avatar  formData  file  true  "Avatar image file"
// @Success      200     {object}  dto.AvatarResponse
// @Failure      400     {object}  httpresponse.Error  "Missing or invalid file"
// @Failure      401     {object}  httpresponse.Error
// @Failure      503     {object}  httpresponse.Error  "Upload service unavailable"
// @Router       /api/v1/users/me/avatar [post]
func (h *Handler) UploadAvatar(c *echo.Context) error {
	if h.uploader == nil {
		return c.JSON(http.StatusServiceUnavailable, httpresponse.NewError(http.StatusServiceUnavailable, "Upload service not configured", ""))
	}

	email, err := claimsEmail(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, httpresponse.NewError(http.StatusUnauthorized, "Unauthorized", err.Error()))
	}

	fileHeader, err := c.FormFile("avatar")
	if err != nil {
		return c.JSON(http.StatusBadRequest, httpresponse.NewError(http.StatusBadRequest, "Avatar file is required", err.Error()))
	}

	if fileHeader.Size > maxAvatarSize {
		return c.JSON(http.StatusBadRequest, httpresponse.NewError(http.StatusBadRequest, "File too large (max 5 MB)", ""))
	}

	mimeType := fileHeader.Header.Get("Content-Type")
	if mimeType != "image/jpeg" && mimeType != "image/png" && mimeType != "image/webp" {
		return c.JSON(http.StatusBadRequest, httpresponse.NewError(http.StatusBadRequest, "Only JPEG, PNG, and WebP images are allowed", ""))
	}

	file, err := fileHeader.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, httpresponse.NewError(http.StatusInternalServerError, "Failed to read file", err.Error()))
	}
	defer file.Close()

	result, err := h.uploader.Upload(c.Request().Context(), file, "zick/avatars")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, httpresponse.NewError(http.StatusInternalServerError, "Upload failed", err.Error()))
	}

	if _, err := h.svc.UpdateAvatarURLByEmail(email, result.URL); err != nil {
		_ = h.uploader.Delete(c.Request().Context(), result.PublicID)
		return c.JSON(http.StatusInternalServerError, httpresponse.NewError(http.StatusInternalServerError, "Failed to save avatar URL", err.Error()))
	}

	return c.JSON(http.StatusOK, dto.AvatarResponse{AvatarURL: result.URL})
}


// RegisterDevice godoc
// @Summary      Register device token
// @Description  Registers or refreshes an FCM (Android) or APNs (iOS) push notification token. Upserts on (user_id, token). Available platforms: IOS, ANDROID.
// @Tags         Devices
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      dto.RegisterDeviceRequest  true  "Device token payload"
// @Success      200      {object}  dto.MessageResponse
// @Failure      400      {object}  httpresponse.Error
// @Failure      401      {object}  httpresponse.Error
// @Failure      500      {object}  httpresponse.Error
// @Router       /api/v1/devices [post]
func (h *Handler) RegisterDevice(c *echo.Context) error {
	email, err := claimsEmail(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, httpresponse.NewError(http.StatusUnauthorized, "Unauthorized", err.Error()))
	}

	userID, err := h.svc.GetUserIDByEmail(email)
	if err != nil {
		return c.JSON(http.StatusNotFound, httpresponse.NewError(http.StatusNotFound, "User not found", err.Error()))
	}

	var req dto.RegisterDeviceRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, httpresponse.NewError(http.StatusBadRequest, "Invalid request body", err.Error()))
	}
	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, httpresponse.NewError(http.StatusBadRequest, "Validation failed", err.Error()))
	}

	if err := h.svc.RegisterDevice(userID, req); err != nil {
		return c.JSON(http.StatusInternalServerError, httpresponse.NewError(http.StatusInternalServerError, "Failed to register device", err.Error()))
	}
	return c.JSON(http.StatusOK, dto.MessageResponse{Message: "Device registered successfully"})
}


// claimsEmail extracts the email from JWT claims stored in Echo context.
// NOTE: The current jwt.go stores email in JwtCustomClaims.Email; UUID is resolved by email.
// A future enhancement should embed the UUID directly into the JWT claims.
func claimsEmail(c *echo.Context) (string, error) {
	claims, ok := c.Get("user").(*auth.JwtCustomClaims)
	if !ok || claims == nil {
		return "", errors.New("missing auth claims")
	}
	return claims.Email, nil
}

// claimsUserID is a convenience wrapper for handlers that just need the UUID.
func claimsUserID(_ *echo.Context, svc Service, email string) (uuid.UUID, error) {
	return svc.GetUserIDByEmail(email)
}

func setTokenCookies(c *echo.Context, accessToken, refreshToken string) {
	c.SetCookie(&http.Cookie{Name: "access_token", Value: accessToken, HttpOnly: true, Path: "/", Expires: time.Now().Add(24 * time.Hour)})
	c.SetCookie(&http.Cookie{Name: "refresh_token", Value: refreshToken, HttpOnly: true, Path: "/", Expires: time.Now().Add(30 * 24 * time.Hour)})
}

func clearTokenCookies(c *echo.Context) {
	c.SetCookie(&http.Cookie{Name: "access_token", Value: "", HttpOnly: true, Path: "/", Expires: time.Unix(0, 0)})
	c.SetCookie(&http.Cookie{Name: "refresh_token", Value: "", HttpOnly: true, Path: "/", Expires: time.Unix(0, 0)})
}

func readRefreshToken(c *echo.Context) string {
	if cookie, err := c.Cookie("refresh_token"); err == nil && cookie.Value != "" {
		return cookie.Value
	}
	var body dto.RefreshRequest
	if err := c.Bind(&body); err == nil {
		return body.RefreshToken
	}
	return ""
}
