package library

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
	v1.GET("/library", h.FindAll)
	v1.GET("/library/:id", h.GetDetails)
	v1.GET("/library-categories", h.FindAllCategories)

	// Admin routes
	adminMW := middlewares.AuthMiddleware(jwtSvc)
	requireAdmin := middlewares.RequireAdmin

	adminGroup := v1.Group("/admin/library", adminMW, requireAdmin)
	adminGroup.POST("", h.Create)
	adminGroup.PUT("/:id", h.Update)
	adminGroup.DELETE("/:id", h.Delete)

	adminCatGroup := v1.Group("/admin/library-categories", adminMW, requireAdmin)
	adminCatGroup.POST("", h.CreateCategory)
	adminCatGroup.PUT("/:id", h.UpdateCategory)
	adminCatGroup.DELETE("/:id", h.DeleteCategory)
}
