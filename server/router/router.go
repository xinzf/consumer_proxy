package router

import (
	"github.com/gin-gonic/gin"
	"github.com/xinzf/consumer_proxy/server/handlers"
	"github.com/xinzf/consumer_proxy/server/router/middleware"
)

func Load(g *gin.Engine) *gin.Engine {

	// 防止 Panic 把进程干死
	g.Use(gin.Recovery())
	g.Use(middleware.Logger())

	// 默认404
	g.NoRoute(func(context *gin.Context) {
		context.JSON(404, gin.H{
			"code": 404,
			"msg":  "请求地址有误，请核实",
			"data": gin.H{},
		})
	})

	g.GET("/check", new(handlers.Home).Check)

	//mq := g.Group("/mq")
	//{
	//	mqHandler := new(handlers.Mq)
	//	mq.GET("/status", mqHandler.Status)
	//	mq.GET("/stop", mqHandler.Stop)
	//	mq.GET("/start", mqHandler.Start)
	//	mq.GET("/remove", mqHandler.Remove)
	//	mq.GET("/reread", mqHandler.Reread)
	//	mq.GET("/update", mqHandler.Update)
	//	mq.GET("/restart", mqHandler.Restart)
	//}
	return g
}
