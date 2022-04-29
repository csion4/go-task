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

func AddWorker(c *gin.Context) {
	var worker vo.AddWorkerReq
	if err := c.ShouldBind(&worker); err != nil {
		panic(err)
	}

	// 埋点
	port, _ := strconv.Atoi(cluster.Track(worker.Ip, worker.UserName, worker.Password))

	// 保存host信息
	db := common.GetDb()
	var workerNode = dto.WorkerNode{Ip: worker.Ip, Port: port, UserName: worker.UserName, Password: worker.Password,
		NodeStatus: 1, CreateTime: time.Now(), CreateUser: 1}
	db.Create(&workerNode)

	response.Success(c, nil, "节点添加成功")
}

func ClusterResp(c *gin.Context)  {
	taskCode := c.Query("taskCode")
	recordId := c.Query("recordId")

	filePath := viper.GetString("taskLog") + taskCode + "/" + taskCode + "_" + recordId + ".log"
	logFile, err2 := os.Create(filePath)
	if err2 != nil {
		panic(err2)
	}
	defer logFile.Close()

	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		panic(err)
	}
	defer ws.Close()

	for {
		_, b, err := ws.ReadMessage()
		if err != nil {
			panic(err)
		}
		if bytes.Equal(b, OverFlag)  {
			return
		}
		_, err = logFile.Write(b)
		if err != nil {
			panic(err)
		}
	}
}

