package middleware

import (
	"flash_drop/helper"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(types string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("token")
		err, claim := helper.VerifyToken(token)
		if err != nil {
			helper.ErrorResponse(c, 401, err.Error(), "You are not authorized")
			c.Abort()
			return
		}

		claims, ok := claim.(jwt.MapClaims)
		if !ok {
			helper.ErrorResponse(c, 401, err, "Something went wrong please try again later")
			c.Abort()
			return
		}

		if claims["role"].(string) != types {
			helper.ErrorResponse(c, 401, err, "You are not authorized")
			c.Abort()
			return
		}

		c.Set("user_id", claims["user_id"])
		c.Next()
		return
	}
}
