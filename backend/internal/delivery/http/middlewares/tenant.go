package middlewares

import (
	"context"
	"net/http"
	"strings"

	"github.com/RianAsmara/cargo-tracker/internal/domain"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func TenantFromHeader(tenantRepo domain.TenantRepository) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			domainStr := c.Request().Header.Get("X-Tenant-Domain")
			if domainStr == "" {
				// Fallback to query param for easier testing
				domainStr = c.QueryParam("tenant_domain")
			}

			if domainStr == "" {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "tenant domain is required"})
			}

			// Get tenant from repo
			tenant, err := tenantRepo.GetByDomain(c.Request().Context(), domainStr)
			if err != nil {
				return c.JSON(http.StatusNotFound, map[string]string{"error": "tenant not found"})
			}

			// Set in request context
			ctx := context.WithValue(c.Request().Context(), domain.TenantContextKey, tenant.ID)
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}

func RequireAdminAuth(jwtSecret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "missing or invalid authorization header"})
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

			token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
				return []byte(jwtSecret), nil
			})

			if err != nil || !token.Valid {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token"})
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token claims"})
			}

			tenantIDStr, ok := claims["tenant_id"].(string)
			if !ok {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "tenant missing from token"})
			}

			tenantID, err := uuid.Parse(tenantIDStr)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid tenant id"})
			}

			// Set in request context
			ctx := context.WithValue(c.Request().Context(), domain.TenantContextKey, tenantID)
			c.SetRequest(c.Request().WithContext(ctx))

			// Set user role
			if roleStr, ok := claims["role"].(string); ok {
				c.Set("user_role", roleStr)
			}

			return next(c)
		}
	}
}

func RequireRole(allowedRoles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userRole, ok := c.Get("user_role").(string)
			if !ok {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized context variable"})
			}

			for _, role := range allowedRoles {
				if userRole == role {
					return next(c)
				}
			}

			return c.JSON(http.StatusForbidden, map[string]string{
				"error": "insufficient permissions",
			})
		}
	}
}
