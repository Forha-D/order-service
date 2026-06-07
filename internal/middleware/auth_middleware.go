package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// type ContextKey string

// const (
// 	UserIDKey ContextKey = "userID"
// 	RoleKey   ContextKey = "role"
// )

func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {

		// 1. Extract from Kong headers
		userID := c.Request().Header.Get("X-User-ID")
		role := c.Request().Header.Get("X-User-Role")

		// 2. Validate presence
		if userID == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "missing user identity",
			})
		}

		if role == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "missing user role",
			})
		}

		// 3. Store in Echo context
		c.Set("userID", userID)
		c.Set("role", role)

		// 4. Continue request
		return next(c)
	}
}

func GetUserID(c echo.Context) string {
	val := c.Get("userID")
	if val == nil {
		return ""
	}
	return val.(string)
}

func GetRole(c echo.Context) string {
	val := c.Get("role")
	if val == nil {
		return ""
	}
	return val.(string)
}

func RequireRole(c echo.Context, allowed string) bool {
	return GetRole(c) == allowed
}
