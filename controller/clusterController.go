package controller

import (
	"bytes"
	"com.csion/tasks/cluster"
	"com.csion/tasks/common"
	"com.csion/tasks/dto"
	"com.csion/tasks/response"
	"com.csion/tasks/vo"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"os"
	"strconv"
	"time"
)

var OverFlag = []byte("Over!")

// 添加工作节点
func AddWorker(c *gin.Context) {
	var worker vo.AddWorkerReq
	log.Panic2("入参异常", c.ShouldBind(&worker))

	// 埋点
	port, _ := strconv.Atoi(cluster.Track(worker.Ip, worker.UserName, worker.Password))

	// 保存host信息
	db := common.GetDb()
	var workerNode = dto.WorkerNode{Ip: worker.Ip, Port: port, UserName: worker.UserName, Password: worker.Password,
		NodeStatus: 1, CreateTime: time.Now(), CreateUser: c.GetInt("userId")}
	log.Panic2("数据操作异常", db.Create(&workerNode).Error)

	response.Success(c, nil, "节点添加成功")
}

// 添加工作节点
func GetWorker(c *gin.Context) {
	name := c.Query("name")

	var list []dto.WorkerNode
	db := common.GetDb().Where("status = 1")
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

	filePath := viper.GetString("taskLog") + taskCode + "/" + taskCode + "_" + recordId + ".log"
	logFile, err2 := os.Create(filePath)
	log.Panic2("工作节点日志会写异常", err2)
	defer logFile.Close()

	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	log.Panic2("工作节点日志会写异常", err)
	defer ws.Close()

	for {
		_, b, err := ws.ReadMessage()
		log.Panic2("工作节点日志会写异常", err)
		if bytes.Equal(b, OverFlag)  {
			return
		}
		_, err = logFile.Write(b)
		log.Panic2("工作节点日志会写异常", err)
	}
}

