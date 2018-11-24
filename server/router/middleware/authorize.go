package middleware

import (
	"github.com/gin-gonic/gin"
)

func Authorize(context *gin.Context) {
	context.Next()
}
