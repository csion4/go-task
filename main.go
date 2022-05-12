package main

import (
	_ "com.csion/tasks/common" // 初始化数据库链接
	_ "com.csion/tasks/config" // 加载配置文件
	"com.csion/tasks/tLog"
	tLogAdapter "com.csion/tasks/tLog/tLogApapter"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

var log = tLog.GetTLog()

// @title task
// @version 1.0
// @description task
// @termsOfService https://github.com/csion4/go-task
// @contact.name csion
// @contact.url https://github.com/csion4
// @contact.email csion4@163.com
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host 127.0.0.1:9044
// @BasePath /
func main() {
	// 初始化环境
	// initEnv()

	// 指定gin的tlog适配
	gin.DefaultWriter = &tLogAdapter.TLogGinAdapter{}
	gin.DefaultErrorWriter = &tLogAdapter.TLogGinAdapter{}
	// gin.SetMode() 调整gin默认日志级别

	log.Debug("启动Web服务")
	// 配置gin
	r := gin.Default()
	r = Route(r)

	// 启动gin
	log.Panic2("gin服务启动异常", r.Run(viper.GetString("server.port")))
}

