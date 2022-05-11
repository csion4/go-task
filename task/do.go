package task

import (
	"com.csion/tasks/common"
	"com.csion/tasks/dto"
	"com.csion/tasks/tLog"
	"com.csion/tasks/utils"
	"github.com/spf13/viper"
	"os"
	"strconv"
)

var chanMap = make(map[string]chan int) // >1:表示各个节点，0：全部完成，-1：当前节点异常了
var log = tLog.GetTLog()

// 构建任务
func RunTask(taskCode string, taskId int, recordId int){
	var stage []dto.TaskStages
	db := common.GetDb()
	log.Debug("任务开始构建，任务编号：", taskCode)

	// 兜底修改任务状态为异常
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
		}
	}()

	// 查询任务节点
	log.Panic2("任务节点查询异常，任务编号：" + taskCode + " ", db.Where("task_id = ? and status =1", taskId).Scan(&stage).Error)

	// ----- before task -----
	// 创建并获取日志文件
	taskLogDir := viper.GetString("taskLog") + taskCode + "/"
	log.Panic2("任务执行目录创建异常，任务编号：" + taskCode + " ", utils.CreateDir(taskLogDir, 0666))

	logFile, err := os.OpenFile(taskLogDir + taskCode + "_" + strconv.Itoa(recordId) + ".log", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	log.Panic2("任务日志文件创建异常，任务编号：" + taskCode + " ", err)
	defer logFile.Close()

	// 初始化目录
	checkErr("初始化工作目录异常", taskCode, utils.CreateDir(viper.GetString("taskWorkspace") + taskCode, 0666), logFile)
	checkErr("初始化工作脚本执行目录异常", taskCode, utils.CreateDir(viper.GetString("taskWorkspace") + taskCode + "@script", 0666), logFile)

	// ----- do stage -----
	for n, value := range stage {
		checkErr("任务节点状态更新异常", taskCode, beforeState(taskId, recordId, value.StageType), logFile)
		env, err := getEnv(value.Id)
		checkErr("获取任务节点异常", taskCode, err, logFile)
		switch value.StageType {
		case 1:
			if _, e := logFile.Write([]byte("----【start stage clone git project】---- \n")); e != nil {
				log.Panic2("日志写入异常，任务编号：" + taskCode, e)
			}
			log.Debug("----【start stage clone git project】----，任务编号：", taskCode)
			Git(env["gitUrl"], env["branch"], viper.GetString("taskWorkspace") + taskCode, logFile)
			break
		case 2:
			if _, e := logFile.Write([]byte("----【start stage exec script】---- \n")); e != nil {
				log.Panic2("日志写入异常，任务编号：" + taskCode, e)
			}
			log.Debug("----【start stage exec script】----，任务编号：", taskCode)
			ExecScript(env["script"], viper.GetString("taskWorkspace") + taskCode + "@script" , viper.GetString("taskWorkspace") + taskCode, logFile)
			break
		case 3:
			// HttpInvoke()
			break
		}
		checkErr("任务节点状态更新异常", taskCode, afterState(taskId, recordId, value.StageType, 2, n), logFile)
	}
	// 0 表示全部完全
	if ch := chanMap[strconv.Itoa(taskId) + strconv.Itoa(recordId)]; ch != nil {
		ch <- 0
	}

	checkErr("任务节点状态更新异常", taskCode, db.Exec("update tasks set task_status = 2 where task_code = ?", taskCode).Error, logFile)
	checkErr("任务节点状态更新异常", taskCode, db.Exec("update task_exec_recode_" + strconv.Itoa(taskId) + " set task_status = 2, update_time = now() where task_status = 1").Error, logFile)
}

// 开始stage执行之前
func beforeState(taskId int, recordId int, stageType int) error {
	// 更新状态
	db := common.GetDb()
	return db.Exec("update task_exec_stage_result_"+strconv.Itoa(taskId)+" set stage_status = 1, create_time = now() "+
		"where record_id = ? and stage_type = ? and stage_status = 0 ORDER by id LIMIT 1", recordId, stageType).Error
}

func afterState(taskId int, recordId int, stageType int, stageStatus int, n int) error {
	if ch := chanMap[strconv.Itoa(taskId) + strconv.Itoa(recordId)]; ch != nil {
		ch <- n + 1
	}
	db := common.GetDb()
	return db.Exec("update task_exec_stage_result_" + strconv.Itoa(taskId) + " set stage_status = ?, update_time = now() " +
		"where record_id = ? and stage_type = ? and stage_status = 1 ORDER by id LIMIT 1", stageStatus, recordId, stageType).Error
}

func getEnv(stageId int) (env map[string]string, err error) {
	var tasksEnv []dto.TaskEnvs
	db := common.GetDb()
	find := db.Where("stage_id = ? and status = 1", stageId).Find(&tasksEnv)
	if find.Error != nil {
		return nil, find.Error
	}

	env = make(map[string]string, len(tasksEnv))
	for _, v := range tasksEnv {
		env[v.Param] = v.Value
	}
	return env, nil
}

// 异常校验，结果写入到执行日志和系统日志中
func checkErr(s string, taskCode string, err error, logFile *os.File) {
	if err != nil {
		_, e := logFile.Write([]byte("【ERROR】 " + s + err.Error() + " \n"))
		log.Panic2("日志写入异常， 任务编号：" + taskCode, e)
		log.Panic2(s + ", 任务编号：" + taskCode, err)
	}
}

func RemoveChan(taskId string, recordId string) {
	ch := chanMap[taskId+recordId]
	delete(chanMap, taskId+recordId) // 清除chanMap
	close(ch)	// 关闭通道
}

