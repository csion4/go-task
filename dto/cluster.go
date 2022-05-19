package dto

import (
	"com.csion/tasks/common"
	"com.csion/tasks/tLog"
	"time"
)

var log = tLog.GetTLog()

type WorkerNode struct {
	Id         int
	Name       string
	NodeStatus int
	Type       int
	Ip         string
	Port       int
	UserName   string
	Password   string
	Strategy   int
	TaskHome   string
	TaskNum    int
	CreateTime time.Time
	UpdateTime time.Time `gorm:"default:null"`
	CreateUser int
	UpdateUser int	`gorm:"default:null"`
}

func (wn *WorkerNode) SelectNode() (r WorkerNode) {
	log.Panic2("数据操作异常", common.GetDb().Where("status = 1 and node_status = 1 and strategy = 0 order by task_num asc limit 1").Find(&r).Error)
	return
}

func (wn *WorkerNode) TaskNumAdd(id int) {
	log.Panic2("数据操作异常", common.GetDb().Exec("update worker_nodes set task_num = task_num + 1 where id = ?", id).Error)
}

func (wn *WorkerNode) TaskNumDec(taskId string, recordId int) {
	log.Panic2("数据操作异常", common.GetDb().Exec("update worker_nodes set task_num = task_num - 1 where id = (select node_id from task_exec_recode_" + taskId + " where id = ?)", recordId).Error)
}

func (wn *WorkerNode) FindAll() (wns []WorkerNode) {
	log.Panic2("数据操作异常", common.GetDb().Where("status = 1").Find(&wns).Error)
	return
}

func (wn *WorkerNode) SaveOne(n *WorkerNode){
	log.Panic2("数据操作异常", common.GetDb().Create(&n).Error)
}