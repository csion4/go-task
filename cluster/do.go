package cluster

import (
	"bytes"
	"com.csion/tasks/common"
	"com.csion/tasks/dto"
	"com.csion/tasks/vo"
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"net/http"
	"os"
)

func DoClusterTask(taskCode string, taskId int, recordId int, worker dto.WorkerNode, logFile *os.File) {
	// 选择worker node
	db := common.GetDb()

	// 封装任务信息
	var taskVO vo.DoClusterTaskVO
	var data []map[int]map[string]string

	var stages []dto.TaskStages
	checkErr("数据操作异常", taskCode, db.Where("task_id = ? and status =1", taskId).Find(&stages).Error, logFile)

	for _, stage := range stages {
		var tasksEnv []dto.TaskEnvs
		checkErr("数据操作异常", taskCode, db.Where("stage_id = ? and status = 1", stage.Id).Find(&tasksEnv).Error, logFile)
		m := make(map[int]map[string]string)
		env := make(map[string]string, len(tasksEnv))
		for _, v := range tasksEnv {
			env[v.Param] = v.Value
		}
		m[stage.StageType] = env
		data = append(data, m)
	}
	taskVO.TaskCode = taskCode
	taskVO.Stages = data
	taskVO.RecordId = recordId

	dataJson, _ := json.Marshal(taskVO)
	//fmt.Println(data2)	// [map[1:map[authType:1 branch:master gitPasswd: gitUrl:http://10.124.192.127/guozh/ttt.git gitUser: token:]] map[2:map[script:mvn clean package]]]

	// 发送任务
	client := &http.Client{}
	request, err := http.NewRequest("POST", fmt.Sprintf("http://%s:%d/task", worker.Ip, worker.Port), bytes.NewBuffer(dataJson))
	checkErr("工作节点任务发起异常", taskCode, err, logFile)
	request.Header.Add("auth", viper.GetString("task.worker.Auth"))
	request.Header.Add("Content-Type", "application/json")

	re, err := client.Do(request)
	checkErr("工作节点任务发起异常", taskCode, err, logFile)
	defer re.Body.Close()

}

// 异常校验，结果写入到执行日志和系统日志中
func checkErr(s string, taskCode string, err error, logFile *os.File) {
	if err != nil {
		_, e := logFile.Write([]byte("【ERROR】 " + s + err.Error() + " \n"))
		log.Panic2("日志写入异常， 任务编号：" + taskCode, e)
		log.Panic2(s + ", 任务编号：" + taskCode, err)
	}
}
