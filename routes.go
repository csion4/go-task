package main

import (
	"com.csion/tasks/controller"
	"com.csion/tasks/middleware"
	"github.com/gin-gonic/gin"
)

// 全局路由配置
func Route(r *gin.Engine) *gin.Engine {

	r.Use(middleware.CORSMiddleware(), middleware.RecoveryMiddleware())

	// 注册登陆
	r.POST("/register", controller.Register)
	r.POST("/login", controller.Login)


	taskGroup := r.Group("/task", middleware.AuthMiddleware())	// 任务创建
	taskGroup.POST("", controller.AddTask) // controller.AddJob就是一个type HandlerFunc func(*Context)类型
	taskGroup.DELETE("", controller.DelTask)
	taskGroup.GET("", controller.GetTasks)


	layoutGroup := r.Group("/taskLayout", middleware.AuthMiddleware()) 	// 任务编排
	layoutGroup.POST("/layoutTask", controller.LayoutTask)
	layoutGroup.GET("/layoutInfo", controller.TaskLayoutInfo)

	runGroup := r.Group("/run", middleware.AuthMiddleware()) 	//构建任务
	runGroup.GET("", controller.RunJob)

	taskRecord := r.Group("/record", middleware.AuthMiddleware()) // 执行记录
	taskRecord.GET("", controller.GetTaskRecord)
	taskRecord.GET("/taskLog", controller.GetTaskLog)


	wsGroup := r.Group("/ws")  // websocket
	wsGroup.GET("/taskLog", controller.GetTaskLogForWS)
	wsGroup.GET("/taskStage", controller.UpdateTaskRecord)


 	r.GET("/hello", middleware.AuthMiddleware(), controller.Hello)

	return r
}
