package server

import (
	"fmt"
	"net/http"
	"time"

	_ "gotickets/docs"
	"gotickets/internal/config"

	"github.com/labstack/echo/v5"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"gorm.io/gorm"
)

// WelcomeResponse defines the shape of the root endpoint.
type WelcomeResponse struct {
	Message     string `json:"message"`
	Version     string `json:"version"`
	Environment string `json:"environment"`
	Status      string `json:"status"`
	DocsURL     string `json:"docs_url"`
}

// HealthResponse is the same shape as WelcomeResponse plus DB diagnostics.
type HealthResponse struct {
	Message     string `json:"message"`
	Version     string `json:"version"`
	Environment string `json:"environment"`
	Status      string `json:"status"`
	DocsURL     string `json:"docs_url"`
	Database    string `json:"database"`
	Timestamp   string `json:"timestamp"`
}

// WelcomeHandler godoc
// @Summary      Welcome
// @Description  Root endpoint — confirms the API is running.
// @Tags         System
// @Produce      json
// @Success      200  {object}  WelcomeResponse
// @Router       / [get]
func WelcomeHandler(cfg *config.Config) echo.HandlerFunc {
	return func(c *echo.Context) error {
		scheme := "http"
		if c.Request().TLS != nil {
			scheme = "https"
		}
		return c.JSON(http.StatusOK, WelcomeResponse{
			Message:     "Welcome to ALTAR API",
			Version:     "1.0.0",
			Environment: cfg.AppEnv,
			Status:      "active",
			DocsURL:     fmt.Sprintf("%s://%s/swagger/index.html", scheme, c.Request().Host),
		})
	}
}

// HealthCheckHandler godoc
// @Summary      Health Check
// @Description  Returns API metadata and database connectivity status.
// @Tags         System
// @Produce      json
// @Success      200  {object}  HealthResponse
// @Router       /health [get]
func HealthCheckHandler(db *gorm.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c *echo.Context) error {
		scheme := "http"
		if c.Request().TLS != nil {
			scheme = "https"
		}

		dbStatus := "up"
		sqlDB, err := db.DB()
		if err != nil || sqlDB.Ping() != nil {
			dbStatus = "down"
		}

		return c.JSON(http.StatusOK, HealthResponse{
			Message:     "Welcome to ALTAR API",
			Version:     "1.0.0",
			Environment: cfg.AppEnv,
			Status:      "active",
			DocsURL:     fmt.Sprintf("%s://localhost:%s/swagger/index.html", scheme, cfg.Port),
			Database:    dbStatus,
			Timestamp:   time.Now().Format(time.RFC3339),
		})
	}
}

// RegisterSwagger mounts the Swagger UI at /swagger/index.html.
func RegisterSwagger(e *echo.Echo) {
	e.GET("/swagger/*", echo.WrapHandler(httpSwagger.Handler()))
}
