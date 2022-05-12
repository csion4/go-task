package config

import (
	"com.csion/tasks/tLog"
	"com.csion/tasks/utils"
	"github.com/spf13/viper"
	"os"
	"strings"
)

var log = tLog.GetTLog()

// 初始化配置文件
func init() {
	log.Debug("开启配置文件加载...")
	// 加载服务配置文件
	workDir, _ := os.Getwd()
	viper.SetConfigName("application")
	viper.SetConfigType("yml")
	viper.AddConfigPath(workDir + "/config")
	log.Panic2("服务配置文件加载异常", viper.ReadInConfig())

	log.Debug("初始化工作目录")
	// 初始化工作目录 taskHome、 taskHome/log、 taskHome/workspace
	taskHome := viper.GetString("task.home")
	if !strings.HasSuffix(taskHome, "/") && !strings.HasSuffix(taskHome, "\\") {
		taskHome += "/"
	}
	log.Panic2("初始化任务日志目录异常：", utils.CreateDir(taskHome + "log/", 0666))
	log.Panic2("初始化任务工作目录异常：", utils.CreateDir(taskHome + "workspace/", 0666))

	viper.Set("taskLog", taskHome + "log/")
	viper.Set("taskWorkspace", taskHome + "workspace/")
	log.Debug("配置文件加载完成！")
}

// 获取工作目录路径
func GetWorkDirPath(taskCode string) (dir string, scriptDir string) {
	dir = viper.GetString("taskWorkspace") + taskCode
	scriptDir = viper.GetString("taskWorkspace") + taskCode + "@script"
	return
}

// 获取日志文件路径
func GetLogFilePath(taskCode string, recordId string) (dir string, file string) {
	dir = viper.GetString("taskLog") + taskCode + "/"
	file = dir + taskCode + "_" + recordId + ".log"
	return
}