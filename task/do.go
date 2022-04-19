package task

import (
	"com.csion/tasks/common"
	"com.csion/tasks/dto"
	"com.csion/tasks/execShell"
	"com.csion/tasks/script"
	"com.csion/tasks/utils"
	"github.com/spf13/viper"
	"log"
	"os"
	"strconv"
)

var chanMap = make(map[string]chan int) // >1:表示各个节点，0：全部完成，-1：当前节点异常了

// 构建任务
func RunTask(taskCode string, taskId int, recordId int){
	var stage []dto.TaskStages
	db := common.GetDb()

	// 兜底修改任务状态为异常
	defer func() {
		r := recover()
		if r != nil {
			db.Exec("update tasks set task_status = 3 where task_code = ?", taskCode)
			db.Exec(`update task_exec_recode_` + strconv.Itoa(taskId) + ` set task_status = 3, update_time = now() where task_status = 1`)
			log.Fatalln(r)
		}
	}()

	// 查询任务节点
	find := db.Raw("select * from task_stages where task_id = ? and status =1", taskId).Scan(&stage)
	if find.Error != nil {
		log.Fatal(find.Error)
	}

	// 创建并获取日志文件
	taskLogDir := viper.GetString("taskLog") + taskCode + "/"
	if err := utils.CreateDir(taskLogDir, 0666); err != nil{
		log.Fatal(err)
	}
	file, err := os.OpenFile(taskLogDir + taskCode + "_" + strconv.Itoa(recordId) + ".log", os.O_CREATE|os.O_APPEND|os.O_SYNC, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		r := recover()
		if r != nil {
			log.Fatal(r)
			//file.Write([]byte(r.(string))) // todo: how write to log file
		}
		_ = file.Close()
	}()

	// 执行stage
	for n, value := range stage {
		beforeState(taskId, recordId, value.StageType, taskCode)
		env := getEnv(value.Id)
		switch value.StageType {
		case 1:
			file.Write([]byte("----【start stage clone git project】---- \n"))
			Git(env["gitUrl"], env["branch"], viper.GetString("taskWorkspace") + taskCode, file)
			break
		case 2:
			file.Write([]byte("----【start stage exec script】---- \n"))
			ExecScript(env["script"], viper.GetString("taskWorkspace") + taskCode + "@script" , viper.GetString("taskWorkspace") + taskCode, file)
			break
		case 3:
			// HttpInvoke()
			break
		default:
			log.Println("unSupport stage type")
		}
		afterState(taskId, recordId, value.StageType, 2, n)  // todo：这边可以做成返回err而不是直接panic的方式，否则无法更行返回状态；从这里思考一下，go用返回err加defer-panic-recover方式来代替try-catch是不是可以完全代替，这里应该怎么保证某个节点失败时的结果记录（带思考）
	}
	// 0 表示全部完全
	if ch := chanMap[strconv.Itoa(taskId) + strconv.Itoa(recordId)]; ch != nil {
		ch <- 0
	}

	db.Exec("update tasks set task_status = 2 where task_code = ?", taskCode)
	db.Exec(`update task_exec_recode_` + strconv.Itoa(taskId) + ` set task_status = 2, update_time = now() where task_status = 1`)
}

// 开始stage执行之前
func beforeState(taskId int, recordId int, stageType int, taskCode string) {
	// 更新状态
	db := common.GetDb()
	db.Exec("update task_exec_stage_result_" + strconv.Itoa(taskId) + " set stage_status = 1, create_time = now() " +
		"where record_id = ? and stage_type = ? and stage_status = 0 ORDER by id LIMIT 1", recordId, stageType)

	// 初始化目录
	if err := utils.CreateDir(viper.GetString("taskWorkspace") + taskCode, 0666); err != nil{
		log.Fatal(err)
	}
	if err := utils.CreateDir(viper.GetString("taskWorkspace") + taskCode + "@script", 0666); err != nil{
		log.Fatal(err)
	}



}

func afterState(taskId int, recordId int, stageType int, stageStatus int, n int) { // todo: 这个暂时所有的结果执行结果都是真确的，不考虑错误的情况
	if ch := chanMap[strconv.Itoa(taskId) + strconv.Itoa(recordId)]; ch != nil {
		ch <- n + 1
	}
	db := common.GetDb()
	db.Exec("update task_exec_stage_result_" + strconv.Itoa(taskId) + " set stage_status = ?, update_time = now() " +
		"where record_id = ? and stage_type = ? and stage_status = 1 ORDER by id LIMIT 1", stageStatus, recordId, stageType)
}

func getEnv(stageId int) (env map[string]string) {

	var tasksEnv []dto.TaskEnvs
	db := common.GetDb()
	find := db.Where("stage_id = ? and status = 1", stageId).Find(&tasksEnv)
	if find.Error != nil {
		log.Fatal(find.Error)
	}

	env = make(map[string]string, len(tasksEnv))
	for _, v := range tasksEnv {
		env[v.Param] = v.Value
	}
	return env
}

// git交互下代码
func Git(url string, branch string, workDir string, file *os.File){
	execShell.ExecShell("git init & git remote add origin "+url, workDir, file)
	execShell.ExecShell("git fetch origin", workDir, file)
	execShell.ExecShell("git checkout -b " + branch + " origin/" + branch, workDir, file)
}

// 执行脚本
func ExecScript(scripts string, scriptDir string, workDir string, file *os.File){
	filePath := script.CreateTempShell(scriptDir, scripts)
	execShell.ExecShell(filePath, workDir, file)
	script.DelFile(filePath)
}

// http调用
func HttpInvoke(url string, param string, t string){

}

func GetTaskStageChan(taskId string, recordId string) (chan int) {
	ch := make(chan int)
	chanMap[taskId + recordId] = ch  // 暂时不用缓冲，缓冲可以避免某些异常，但是可能导致gc失败，暂时不用缓冲
	return ch
}

func RemoveChan(taskId string, recordId string) {
	ch := chanMap[taskId+recordId]
	delete(chanMap, taskId+recordId) // 清除chanMap
	close(ch)	// 关闭通道，会报错吗？
}

