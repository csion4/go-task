package task

import (
	"com.csion/tasks/cluster"
	"com.csion/tasks/common"
	"com.csion/tasks/config"
	"com.csion/tasks/dto"
	"com.csion/tasks/utils"
	"os"
	"strconv"
	"sync"
)

var lock sync.Mutex

// 任务分发
func PublishTask(taskCode string, taskId int, recordId int) {
	// todo: 这里任务状态需要事务，保证数据一致性，因为时分库分表，所以不用担心锁表和效率问题
	// 兜底修改任务状态为异常
	db := common.GetDb()
	defer func() {
		r := recover()
		if r != nil {
			r := db.Exec("update tasks set task_status = 3 where task_code = ?", taskCode)
			if r.Error != nil {
				log.Error("任务<", taskCode, ">回退数据库执行异常", r.Error)
			}
			r2 := db.Exec(`update task_exec_recode_` + strconv.Itoa(taskId) + ` set task_status = 3, update_time = now() where task_status = 1`)
			if r2.Error != nil {
				log.Error("任务<", taskCode, ">回退数据库执行异常", r.Error)
			}
			log.Error("任务构建异常，设置执行状态失败，任务编号：", taskCode)
			finishNode(taskId, recordId)
		}
	}()

	node := selectNode()
	db.Exec("update task_exec_recode_" + strconv.Itoa(taskId) + " set node_id = ? where id = ?", node.Id, recordId)
	logFile := createLog(taskCode, recordId)
	defer logFile.Close()
	_, err := logFile.Write([]byte("---- 【select node】 ---- \n【node】此次编译节点：" + node.Name + "\n" ))
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
	wn = wn.SelectNode()
	wn.TaskNumAdd(wn.Id)
	return
}

// 选择节点
func finishNode(taskId int, recordId int){
	lock.Lock()
	defer lock.Unlock()
	var wn dto.WorkerNode
	wn.TaskNumDec(strconv.Itoa(taskId), recordId)
	return
}

func createLog(taskCode string, recordId int) *os.File {
	logDir, file := config.GetLogFilePath(taskCode, strconv.Itoa(recordId))
	log.Panic2("任务执行目录创建异常，任务编号：" + taskCode + " ", utils.CreateDir(logDir, 0666))
	logFile, err := os.OpenFile(file, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	log.Panic2("任务日志文件创建异常，任务编号：" + taskCode + " ", err)
	return logFile
}