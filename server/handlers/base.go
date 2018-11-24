package handlers

import (
	"github.com/gin-gonic/gin"

	"github.com/xinzf/consumer_proxy/server/pkg/errno"
	"net/http"
)

type Base struct {
}

// Success 执行成功
func (this *Base) Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"msg_code":   errno.OK.Code,
		"message":    errno.OK.Message,
		"attachment": data,
	})
}

// Failed 执行失败
func (this *Base) Failed(c *gin.Context, code *errno.Errno) {
	c.JSON(http.StatusOK, gin.H{
		"msg_code":   code.Code,
		"message":    code.Message,
		"attachment": gin.H{},
	})
}
