package main

import (
	"com.csion/tasks/cluster"
	"com.csion/tasks/common"
	"com.csion/tasks/dto"
	"github.com/spf13/viper"
	"gorm.io/gorm"
	"strconv"
	"time"
)

// 在服务启动时，初始化环境
func initEnv() {
	log.Debug("开始初始化环境...")
	// 清理任务状态
	db := common.GetDb()
	_ = db.Transaction(func(tx *gorm.DB) error {
		var tasks []dto.Tasks
		log.Panic2("数据操作异常", tx.Where("status = 1 and task_status = 1").Find(&tasks).Error)
		for _, task := range tasks {
			log.Panic2("数据操作异常", tx.Exec("update task_exec_recode_" + strconv.Itoa(task.Id) + " set task_status = 3 where task_status = 1").Error)
			log.Panic2("数据操作异常", tx.Exec("update task_exec_stage_result_" + strconv.Itoa(task.Id) + " set stage_status = 3 where stage_status in (1, 0)").Error)
		}
		if len(tasks) > 0 {
			log.Panic2("数据操作异常", tx.Exec("update tasks set task_status = 3 where task_status = 1").Error)
		}
		return nil
	})

	// 清理节点任务数
	db.Exec("update worker_nodes set task_num = 0 where status = 1")

	// 初始化worker节点
	var wn dto.WorkerNode
	all := wn.FindAll()
	var flag bool
	for _, node := range all {
		if node.Type == 1 {
			port := checkNode(node.Ip, node.UserName, node.Password)
			if port == "" && node.NodeStatus == 1 {
				db.Exec("update worker_nodes set node_status = 2 where id = ?", node.Id)
			} else if port != ""{
				p, _ := strconv.Atoi(port)
				if p != node.Port {
					db.Exec("update worker_nodes set port = ? where id = ?", p, node.Id)
				}
			}
			cluster.NodeProbe(node.Id, node.Ip, node.Port)
		} else {
			flag = true
		}
	}

	// 注册自身节点
	if !flag {
		masterNode := dto.WorkerNode{
			Name:       "master",
			NodeStatus: 1,
			Type:       0,
			Ip:         "0.0.0.0",
			Port:       0,
			UserName:   "",
			Password:   "",
			Strategy:   0,
			TaskHome:   viper.GetString("task.home"),
			CreateTime: time.Now(),
			CreateUser: 1,
		}
		masterNode.SaveOne(&masterNode)
	}

	log.Debug("环境初始化完成！")
}

func checkNode(ip string, userName string, password string) string {
	defer func() {
		if err := recover();err != nil {
			return
		}
	}()
	return cluster.Track(ip, userName, password)
}