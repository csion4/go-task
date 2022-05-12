package task

import (
	"com.csion/tasks/cluster"
	"com.csion/tasks/config"
	"com.csion/tasks/dto"
	"com.csion/tasks/utils"
	"os"
	"strconv"
	"sync"
)

var lock sync.Mutex

func PublishTask(taskCode string, taskId int, recordId int) {
	node := selectNode()
	logFile := createLog(taskCode, recordId)
	defer logFile.Close()
	_, err := logFile.Write([]byte("【SELECTNODE】 此次编译节点：" + node.Name + "\n" ))
	log.Panic2("日志写入异常， 任务编号：" + taskCode, err)
	if node.Type == 0 {
		RunTask(taskCode, taskId , recordId, logFile)
	} else {
		cluster.DoClusterTask(taskCode, taskId , recordId, node, logFile)
	}
}

// 选择节点
func selectNode() (wn dto.WorkerNode) {
	lock.Lock()
	defer lock.Unlock()
	wn = wn.FindByTaskNumAsc()
	wn.TaskNumAdd(wn.Id)
	return
}

func createLog(taskCode string, recordId int) *os.File {
	logDir, file := config.GetLogFilePath(taskCode, strconv.Itoa(recordId))
	log.Panic2("任务执行目录创建异常，任务编号：" + taskCode + " ", utils.CreateDir(logDir, 0666))
	logFile, err := os.OpenFile(file, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	log.Panic2("任务日志文件创建异常，任务编号：" + taskCode + " ", err)
	return logFile
}