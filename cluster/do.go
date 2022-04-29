package cluster

import (
	"bytes"
	"com.csion/tasks/common"
	"com.csion/tasks/dto"
	"com.csion/tasks/vo"
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"log"
	"net/http"
)

func DoClusterTask(taskCode string, taskId int, recordId int) {
	// 选择worker node
	db := common.GetDb()
	var worker dto.WorkerNode
	r := db.First(&worker)	// 具体节点选取规则再定
	if r.Error != nil {
		panic(r.Error)
	}

	// 封装任务信息
	var taskVO vo.DoClusterTaskVO
	var data []map[int]map[string]string

	var stages []dto.TaskStages
	find := db.Where("task_id = ? and status =1", taskId).Find(&stages)
	if find.Error != nil {
		log.Fatal(find.Error)
	}

	for _, stage := range stages {
		var tasksEnv []dto.TaskEnvs
		find := db.Where("stage_id = ? and status = 1", stage.Id).Find(&tasksEnv)
		if find.Error != nil {
			log.Fatal(find.Error)
		}
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
	if err != nil {
		log.Fatal(err)
	}
	request.Header.Add("auth", viper.GetString("task.worker.Auth"))
	request.Header.Add("Content-Type", "application/json")

	re, err := client.Do(request)
	if err != nil {
		log.Fatal(err)
	}
	defer re.Body.Close()

}
