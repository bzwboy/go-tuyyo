package router

import (
	"controller"
	"github.com/gin-gonic/gin"
)

func SetRouter() *gin.Engine {
	r := gin.Default()

/*	v1 := r.Group("/v1")
	{
		// 发送组消息
		v1.POST("/send_message", controller.SendMessageV1)
	}*/

	r.GET("/ping", controller.Ping)

	return r
}
