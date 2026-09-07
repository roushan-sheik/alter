package middlewares

import (
	"gotickets/internal/auth"
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"
)

func AuthMiddleware(jwtService auth.JWTService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			var tokenString string

			authHeader := c.Request().Header.Get("Authorization")
			if authHeader != "" {
				// Automatically support tokens with or without the "Bearer " prefix
				if strings.HasPrefix(authHeader, "Bearer ") {
					tokenString = strings.TrimPrefix(authHeader, "Bearer ")
				} else {
					tokenString = authHeader
				}
			} else if cookie, err := c.Cookie("access_token"); err == nil {
				tokenString = cookie.Value
			}

			if tokenString == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Missing Authorization header or cookie"})
			}
			claims, err := jwtService.ValidateToken(tokenString, false)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Invalid or expired token"})
			}

			// Add the claims to context so handlers can access the user info
			c.Set("user", claims)
			c.Set("is_premium", claims.IsPremium)

			return next(c)
		}
	}
}

// RequirePremium is a middleware that gates routes to premium users only.
// It must be applied after AuthMiddleware (so the "user" claims are already in context).
func RequirePremium(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		// The service layer sets is_premium on the claims context key "is_premium"
		// when resolveIsPremium is called. Alternatively, check the claims directly.
		// For a simple gate, we rely on the service layer's check in the handler —
		// this middleware serves as an explicit layer for future use on route-level enforcement.
		isPremium, _ := c.Get("is_premium").(bool)
		if !isPremium {
			return c.JSON(http.StatusForbidden, map[string]string{
				"code":        "403",
				"message":     "This resource requires a premium subscription.",
				"upgrade_url": "/api/v1/subscriptions/plans",
			})
		}
		return next(c)
	}
}

// RequireAdmin is a middleware that gates routes to admin users only.
// It must be applied after AuthMiddleware (so the "user" claims are already in context).
func RequireAdmin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		claims, ok := c.Get("user").(*auth.JwtCustomClaims)
		if !ok || (claims.Role != "admin" && claims.Role != "ADMIN") {
			return c.JSON(http.StatusForbidden, map[string]string{
				"code":    "403",
				"message": "This resource requires admin privileges.",
			})
		}
		return next(c)
	}
}
