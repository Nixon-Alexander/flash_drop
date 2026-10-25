package middleware

import (
	"flash_drop/helper"
	"flash_drop/types"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(requiredRole types.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, ok := strings.CutPrefix(c.GetHeader("Authorization"), "Bearer ")
		fmt.Println(tokenString)
		if !ok {
			helper.ErrorResponse(c, http.StatusUnauthorized, nil, "Missing or malformed Authorization header")
			c.Abort()
			return
		}

		claims, err := helper.VerifyToken(tokenString)
		if err != nil {
			helper.ErrorResponse(c, http.StatusUnauthorized, nil, "Invalid or expired token")
			c.Abort()
			return
		}

		if claims.Role != requiredRole {
			helper.ErrorResponse(c, http.StatusForbidden, nil, "Forbidden")
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Next()
	}
}
