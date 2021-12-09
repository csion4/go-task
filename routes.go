package main

import (
	"com.csion/tasks/controller"
	"github.com/gin-gonic/gin"
)

// 全局路由配置
func Route(r *gin.Engine) *gin.Engine {
	taskGroup := r.Group("/task")	// 任务创建
	taskGroup.POST("", controller.AddJob)	// controller.AddJob就是一个type HandlerFunc func(*Context)类型
	taskGroup.DELETE("", controller.DelJob)
	taskGroup.GET("", controller.GetJobs)


	layoutGroup := r.Group("/taskLayout") 	// 任务编排
	layoutGroup.POST("/layoutTask", controller.LayoutTask)
	layoutGroup.GET("/layoutInfo", controller.LayoutInfo)

	runGroup := r.Group("/run") 	//构建任务
	runGroup.GET("", controller.RunJob)

	r.GET("/hello", controller.Hello)

	return r
}
