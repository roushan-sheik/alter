package motivation

import (
	"gotickets/internal/auth"
	"gotickets/internal/middlewares"

	"github.com/labstack/echo/v5"
	"gorm.io/gorm"
)

func RegisterRoutes(e *echo.Echo, db *gorm.DB, jwtSvc auth.JWTService) {
	repo := NewRepository(db)
	svc := NewService(repo)
	h := NewHandler(svc)

	v1 := e.Group("/api/v1")

	// Public routes
	v1.GET("/motivations", h.FindAll)
	v1.GET("/motivations/:id", h.GetDetails)

	// Admin routes
	adminMW := middlewares.AuthMiddleware(jwtSvc)
	requireAdmin := middlewares.RequireAdmin

	adminGroup := v1.Group("/admin/motivations", adminMW, requireAdmin)
	adminGroup.POST("", h.Create)
	adminGroup.PUT("/:id", h.Update)
	adminGroup.DELETE("/:id", h.Delete)
}
