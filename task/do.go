package task

import (
	"com.csion/tasks/common"
	"com.csion/tasks/config"
	"com.csion/tasks/dto"
	"com.csion/tasks/tLog"
	"com.csion/tasks/utils"
	"github.com/spf13/viper"
	"os"
	"strconv"
)

var chanMap = make(map[string]chan int) // >1:表示各个节点，0：全部完成，-1：当前节点异常了
var log = tLog.GetTLog()

// 构建master节点任务
func RunTask(taskCode string, taskId int, recordId int, logFile *os.File){
	var stage []dto.TaskStages
	db := common.GetDb()
	log.Debug("任务开始构建，任务编号：", taskCode)

	// 查询任务节点
	log.Panic2("任务节点查询异常，任务编号：" + taskCode + " ", db.Where("task_id = ? and status =1", taskId).Find(&stage).Error)

	// ----- before task -----
	// 创建并获取日志文件
	if logFile == nil {
		logFile = createLog(taskCode, recordId)
		defer logFile.Close()
	}


	// 初始化目录
	workDir, scriptDir := config.GetWorkDirPath(taskCode)
	checkErr("初始化工作目录异常", taskCode, utils.CreateDir(workDir, 0666), logFile)
	checkErr("初始化工作脚本执行目录异常", taskCode, utils.CreateDir(scriptDir, 0666), logFile)

	// ----- do stage -----
	for n, value := range stage {
		checkErr("任务节点状态更新异常", taskCode, BeforeState(taskId, recordId, value.StageType), logFile)
		env, err := getEnv(value.Id)
		checkErr("获取任务节点异常", taskCode, err, logFile)
		switch value.StageType {
		case 1:
			if _, e := logFile.Write([]byte("\n----【start stage clone git project】---- \n")); e != nil {
				log.Panic2("日志写入异常，任务编号：" + taskCode, e)
			}
			log.Debug("----【start stage clone git project】----，任务编号：", taskCode)
			Git(env["gitUrl"], env["branch"], viper.GetString("taskWorkspace") + taskCode, logFile)
			break
		case 2:
			if _, e := logFile.Write([]byte("\n----【start stage exec script】---- \n")); e != nil {
				log.Panic2("日志写入异常，任务编号：" + taskCode, e)
			}
			log.Debug("----【start stage exec script】----，任务编号：", taskCode)
			ExecScript(env["script"], viper.GetString("taskWorkspace") + taskCode + "@script" , viper.GetString("taskWorkspace") + taskCode, logFile)
			break
		case 3:
			// HttpInvoke()
			break
		}
		checkErr("任务节点状态更新异常", taskCode, AfterState(taskId, recordId, value.StageType, 2, n), logFile)
	}

	checkErr("任务节点状态更新异常", taskCode, Success(taskId, recordId, taskCode), logFile)
}

// 开始stage执行之前
func BeforeState(taskId int, recordId int, stageType int) error {
	// 更新状态
	db := common.GetDb()
	return db.Exec("update task_exec_stage_result_" + strconv.Itoa(taskId) + " set stage_status = 1, create_time = now() "+
		"where record_id = ? and stage_type = ? and stage_status = 0 ORDER by id LIMIT 1", recordId, stageType).Error
}

func AfterState(taskId int, recordId int, stageType int, stageStatus int, n int) error {
	if ch := chanMap[strconv.Itoa(taskId) + strconv.Itoa(recordId)]; ch != nil {
		ch <- n + 1
	}
	db := common.GetDb()
	return db.Exec("update task_exec_stage_result_" + strconv.Itoa(taskId) + " set stage_status = ?, update_time = now() " +
		"where record_id = ? and stage_type = ? and stage_status = 1 ORDER by id LIMIT 1", stageStatus, recordId, stageType).Error
}

func Success(taskId int, recordId int, taskCode string) error {
	// 0 表示全部完全
	if ch := chanMap[strconv.Itoa(taskId) + strconv.Itoa(recordId)]; ch != nil {
		ch <- 0
	}

	finishNode(taskId, recordId)
	db := common.GetDb()
	err := db.Exec("update tasks set task_status = 2 where task_code = ?", taskCode).Error
	if err != nil {
		return err
	}
	return db.Exec("update task_exec_recode_" + strconv.Itoa(taskId) + " set task_status = 2, update_time = now() where task_status = 1").Error



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

// 通过一个公共的通道响应执行状态变更
func GetTaskStageChan(taskId string, recordId string) chan int {
	ch := make(chan int)
	chanMap[taskId + recordId] = ch  // 暂时不用缓冲，缓冲可以避免某些异常，但是可能导致gc失败，暂时不用缓冲
	return ch
}
func RemoveChan(taskId string, recordId string) {
	ch := chanMap[taskId+recordId]
	delete(chanMap, taskId+recordId) // 清除chanMap
	close(ch)	// 关闭通道
}

