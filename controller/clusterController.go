package controller

import (
	"com.csion/tasks/cluster"
	"com.csion/tasks/common"
	"com.csion/tasks/config"
	"com.csion/tasks/dto"
	"com.csion/tasks/response"
	"com.csion/tasks/task"
	"com.csion/tasks/vo"
	"github.com/gin-gonic/gin"
	"os"
	"strconv"
	"strings"
	"time"
)

const FailedFlag = "Failed"
const SuccessFlag = "Success"
const StageFlag = "StageFlag"
const StageBefore = "Before"
const StageAfter = "After"

// 添加工作节点
func AddWorker(c *gin.Context) {
	var worker vo.AddWorkerReq
	log.Panic2("入参异常", c.ShouldBind(&worker))

	// 埋点
	port, _ := strconv.Atoi(cluster.Track(worker.Ip, worker.UserName, worker.Password, worker.TaskHome))

	// 保存host信息
	db := common.GetDb()
	var workerNode = dto.WorkerNode{
		Name: worker.Name,
		NodeStatus: 1,
		Type: 1,
		Ip: worker.Ip,
		Port: port,
		UserName: worker.UserName,
		Password: worker.Password,
		Strategy: worker.Strategy,
		TaskHome: worker.TaskHome,
		TaskNum: 0,
		CreateTime: time.Now(),
		CreateUser: c.GetInt("userId")}
	log.Panic2("数据操作异常", db.Create(&workerNode).Error)

	cluster.NodeProbe(workerNode.Id)

	response.Success(c, nil, "节点添加成功")
}

// 查询工作节点
func GetWorker(c *gin.Context) {
	name := c.Query("name")

	var list []dto.WorkerNode
	db := common.GetDb().Select("name, node_status, type, ip, user_name, '*******' as password, strategy, work_home").Where("status = 1")
	if name != "" {
		db = db.Where("name like concat('%', ?, '%') or ip like concat('%', ?, '%')", name, name)
	}
	log.Panic2("数据操作异常", db.Find(&list).Error)

	response.Success(c, gin.H{"data": list}, "查询成功")
}



// master与worker建立ws连接获取响应
func ClusterResp(c *gin.Context)  {
	taskCode := c.Query("taskCode")
	recordId := c.Query("recordId")

	_, filePath := config.GetLogFilePath(taskCode, recordId)
	logFile, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	log.Panic2("任务日志文件创建异常，任务编号：" + taskCode + " ", err)
	defer logFile.Close()

	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	log.Panic2("工作节点日志会写异常", err)
	defer ws.Close()

	var t dto.Tasks
	taskId := t.FindIdFromTaskCode(taskCode)

	for {
		_, b, err := ws.ReadMessage()
		log.Panic2("工作节点日志会写异常", err)
		if strings.HasPrefix(string(b), StageFlag) {
			ri, _ := strconv.Atoi(recordId)
			dealStage(string(b), taskId, ri)
		} else {
			if string(b) == SuccessFlag  {
				ri, _ := strconv.Atoi(recordId)
				taskSuccess(taskId, ri, taskCode)
				return
			}
			if string(b) == FailedFlag {
				taskFailed()
				return
			}
			_, err = logFile.Write(b)
			log.Panic2("工作节点日志会写异常", err)
		}
	}
}

func dealStage(stageInfo string, taskId int, recordId int) {
	s := strings.Split(stageInfo, "-")
	stageType, _ := strconv.Atoi(s[2])
	if s[1] == StageBefore {
		err := task.BeforeState(taskId, recordId, stageType)
		if err != nil {
			log.Error(err)
		}
	} else if s[1] == StageAfter {
		n, _ := strconv.Atoi(s[3])
		err := task.AfterState(taskId, recordId, stageType, 2, n)
		if err != nil {
			log.Error(err)
		}
	}
}

func taskSuccess(taskId int, recordId int, taskCode string) {
	_ = task.Success(taskId, recordId, taskCode)
}

// todo:====
func taskFailed() {

}

// 响应node节点的监听
func ClusterPing(c *gin.Context)  {
	response.Success(c, gin.H{"data": "Pong"}, "Pong")
}

