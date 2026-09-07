package user

import (
	"context"
	"gotickets/internal/auth"
	"gotickets/internal/config"
	"gotickets/internal/email"
	"gotickets/internal/middlewares"
	"gotickets/internal/upload"

	firebase "firebase.google.com/go/v4"
	firebaseAuth "firebase.google.com/go/v4/auth"
	"github.com/labstack/echo/v5"
	"gorm.io/gorm"
)

func RegisterRoutes(e *echo.Echo, db *gorm.DB, cfg *config.Config, uploader upload.Uploader) {
	repo := NewRepository(db)
	jwtSvc := auth.NewJWTService(cfg.JwtAccessSecret, cfg.JwtRefreshSecret, cfg.JwtAccessExpiry, cfg.JwtRefreshExpiry)

	var mailer email.Mailer
	if cfg.SMTPFromAddress != "" && cfg.SMTPAppPassword != "" {
		mailer = email.NewGmailMailer(cfg.SMTPFromName, cfg.SMTPFromAddress, cfg.SMTPAppPassword)
	}

	var fbAuthClient *firebaseAuth.Client
	ctx := context.Background()
	var app *firebase.App
	var err error
	if cfg.FirebaseProjectID != "" {
		app, err = firebase.NewApp(ctx, &firebase.Config{ProjectID: cfg.FirebaseProjectID})
	} else {
		app, err = firebase.NewApp(ctx, nil)
	}
	if err == nil && app != nil {
		fbAuthClient, _ = app.Auth(ctx)
	}

	svc := NewService(repo, jwtSvc, uploader, mailer, fbAuthClient, cfg.AdminEmail, cfg.AdminPassword)
	h := NewHandler(svc, uploader)

	authMW := middlewares.AuthMiddleware(jwtSvc)

	authGroup := e.Group("/api/v1/auth")
	authGroup.POST("/register", h.Register)
	authGroup.POST("/login", h.Login)
	authGroup.POST("/admin/login", h.AdminLogin)
	authGroup.POST("/social-login", h.SocialLogin)
	authGroup.POST("/refresh", h.Refresh)
	authGroup.POST("/logout", h.Logout, authMW)
	authGroup.POST("/forgot-password", h.ForgotPassword)
	authGroup.POST("/resend-otp", h.ResendOTP)
	authGroup.POST("/verify-otp", h.VerifyOTP)
	authGroup.POST("/reset-password", h.ResetPassword)

	userGroup := e.Group("/api/v1/users", authMW)
	userGroup.GET("/me", h.GetMe)
	userGroup.PUT("/me", h.UpdateMe)
	userGroup.PUT("/me/password", h.ChangePassword)
	userGroup.DELETE("/me", h.DeleteMe)
	userGroup.POST("/me/avatar", h.UploadAvatar)

	deviceGroup := e.Group("/api/v1/devices", authMW)
	deviceGroup.POST("", h.RegisterDevice)
}
