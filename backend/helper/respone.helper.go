package helper

import (
	"flash_drop/types"

	"github.com/gin-gonic/gin"
)

func SuccessResponse(c *gin.Context, httpCode int, data any, message string) {
	c.JSON(httpCode, types.ResponseType{
		Status:  "success",
		Code:    httpCode,
		Data:    data,
		Message: message,
	})
}

func ErrorResponse(c *gin.Context, httpCode int, data any, message string) {
	c.JSON(httpCode, types.ResponseType{
		Status:  "error",
		Code:    httpCode,
		Data:    data,
		Message: message,
	})
}
