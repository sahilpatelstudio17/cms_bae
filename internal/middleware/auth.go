package middleware

import (
	"net/http"
	"strings"

	"cms/internal/utils"

	"github.com/gin-gonic/gin"
)

const (
	ContextUserIDKey    = "userID"
	ContextCompanyIDKey = "companyID"
	ContextRoleKey      = "role"
)

func JWTAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.Error(c, http.StatusUnauthorized, "authorization header is required")
			c.Abort()
			return
		}

		tokenString := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		if tokenString == authHeader {
			utils.Error(c, http.StatusUnauthorized, "invalid authorization header format")
			c.Abort()
			return
		}

		claims, err := utils.ParseJWT(tokenString, secret)
		if err != nil {
			utils.Error(c, http.StatusUnauthorized, "invalid or expired token")
			c.Abort()
			return
		}

		c.Set(ContextUserIDKey, claims.UserID)
		c.Set(ContextCompanyIDKey, claims.CompanyID)
		c.Set(ContextRoleKey, claims.Role)
		c.Next()
	}
}

func RequireRoles(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]bool, len(roles))
	for _, role := range roles {
		allowed[role] = true
	}

	return func(c *gin.Context) {
		roleAny, ok := c.Get(ContextRoleKey)
		if !ok {
			utils.Error(c, http.StatusForbidden, "role not found in context")
			c.Abort()
			return
		}

		role, ok := roleAny.(string)
		if !ok || !allowed[role] {
			utils.Error(c, http.StatusForbidden, "insufficient role permissions")
			c.Abort()
			return
		}

		c.Next()
	}
}

func UserIDFromContext(c *gin.Context) (uint, bool) {
	value, ok := c.Get(ContextUserIDKey)
	if !ok {
		return 0, false
	}
	userID, ok := value.(uint)
	return userID, ok
}

func CompanyIDFromContext(c *gin.Context) (uint, bool) {
	value, ok := c.Get(ContextCompanyIDKey)
	if !ok {
		return 0, false
	}
	companyID, ok := value.(uint)
	return companyID, ok
}
