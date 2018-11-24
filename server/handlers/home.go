package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/xinzf/consumer_proxy/server/pkg/errno"
)

type Home struct {
	Base
}

func (this *Home) Check(c *gin.Context) {
	this.Success(c, errno.OK)
}
