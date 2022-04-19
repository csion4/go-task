package main

import (
	_ "com.csion/tasks/common" // 初始化数据库链接
	_ "com.csion/tasks/config" // 加载配置文件
	"com.csion/tasks/utils"
	"github.com/gin-gonic/gin"
	"strings"

	// "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

//// 日志组件
//var log = logrus.New()
//
//func init() {
//	// Log as JSON instead of the default ASCII formatter.
//	log.Formatter = &logrus.JSONFormatter{}
//	// Output to stdout instead of the default stderr
//	// Can be any io.Writer, see below for File example
//	f, _ := os.Create("./gin.log")
//	log.Out = f
//	gin.SetMode(gin.DebugMode)
//	gin.DefaultWriter = log.Out
//	// Only log the warning severity or above.
//	log.Level = logrus.DebugLevel
//	log.Debug("111111111111")
//	log.Errorln()
//}

func init() {
	// 初始化工作目录 taskHome、 taskHome/log、 taskHome/workspace
	taskHome := viper.GetString("task.home")
	if !strings.HasSuffix(taskHome, "/") && !strings.HasSuffix(taskHome, "\\") {
		taskHome += "/"
	}
	if err := utils.CreateDir(taskHome + "log/", 0666); err != nil{
		panic(err)
	}
	if err := utils.CreateDir(taskHome + "workspace/", 0666); err != nil{
		panic(err)
	}

	viper.Set("taskLog", taskHome + "log/")
	viper.Set("taskWorkspace", taskHome + "workspace/")


}


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
	// 配置gin
	r := gin.Default()
	r = Route(r)
	// 启动gin
	panic(r.Run(viper.GetString("server.port")))

}
