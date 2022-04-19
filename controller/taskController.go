package controller

import (
	"bufio"
	"com.csion/tasks/common"
	"com.csion/tasks/dto"
	"com.csion/tasks/response"
	"com.csion/tasks/task"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize: 10,
	WriteBufferSize: 10,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}


// hello
func Hello(c *gin.Context){
	response.Success(c, nil, "hello")
}

// 添加任务
func AddTask(c *gin.Context){
	// 绑定参数
	var tasks dto.Tasks
	err := c.ShouldBind(&tasks)
	if err != nil {
		panic(err)
	}

	// 存表
	var r string
	db := common.GetDb()
	if tasks.Id == 0 {
		tasks.CreateTime = time.Now()
		tasks.CreateUser = 1
		db = db.Create(&tasks)
		r = "添加成功"
	} else {
		tasks.UpdateTime = time.Now()
		tasks.UpdateUser = 1
		db = db.Model(&tasks).Updates(&tasks)
		r = "修改成功"
	}
	if db.Error != nil {
		panic(db.Error)
	}
	response.Success(c, gin.H{"data": tasks}, r)
}

// 删除任务
func DelTask(c *gin.Context){
	taskId := c.Query("taskId")

	db := common.GetDb()
	result := db.Exec("update tasks set status = 0 where id = ?", taskId)
	if result.Error != nil {
		panic(result.Error)
	}

	response.Success(c, nil, "删除成功")
}

//  查询任务
func GetTasks(c *gin.Context){
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))

	var tasks []dto.Tasks
	db := common.GetDb()
	var total int64
	result := db.Model(&dto.Tasks{}).Where("status = ?", 1).Count(&total)
	if result.Error != nil {
		panic(result.Error)
	}
	result = db.Order("id desc").Limit(pageSize).Offset((page - 1) * pageSize).Where("status = ?", 1).Find(&tasks)
	if result.Error != nil {
		panic(result.Error)
	}

	response.Success(c, gin.H{"data": tasks, "total": total}, "查询成功")
}


// 构建任务
func RunJob(c *gin.Context){
	taskCode := c.Query("taskCode")

	// 根据taskCode查找task信息
	db := common.GetDb()
	var taskDto dto.Tasks
	find := db.Where("task_code = ? and status =1", taskCode).Find(&taskDto)
	if find.Error != nil {
		panic(find.Error)
	}

	// 通过判断task的状态判断是否需要初始化任务执行记录表
	if taskDto.TaskStatus == 0 {
		exec := db.Exec(`create table task_exec_recode_` + strconv.Itoa(taskDto.Id) + ` (
    id int(11) NOT NULL AUTO_INCREMENT COMMENT '主键',
	task_status tinyint(1) NOT NULL DEFAULT '1' COMMENT '任务执行状态：1：执行中，2：执行成功，3：执行失败',
    create_time datetime DEFAULT NULL COMMENT '创建时间',
    update_time datetime DEFAULT NULL COMMENT '更新时间',
    PRIMARY KEY (id)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8 COMMENT= '任务执行记录表'`)
		if exec.Error != nil {
			panic(exec.Error)
		}
		exec = db.Exec(`create table task_exec_stage_result_` + strconv.Itoa(taskDto.Id) + ` (
  id int(11) NOT NULL AUTO_INCREMENT COMMENT '主键',
  record_id int(11) NOT NULL COMMENT '执行记录标识',
  stage_type tinyint(1) NOT NULL COMMENT '节点类型',
  stage_status tinyint(1) NOT NULL DEFAULT '1' COMMENT '节点执行状态：1：执行中，2：执行成功，3：执行失败',
  create_time datetime DEFAULT NULL COMMENT '创建时间',
  update_time datetime DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='任务节点执行记录表'`)
		if exec.Error != nil {
			panic(exec.Error)
		}
	}

	// 更新任务状态
	result := db.Exec("update tasks set task_status = 1 where task_code = ?", taskCode)
	if result.Error != nil {
		panic(result.Error)
	}

	// 向任务执行记录表中插入数据
	db.Exec(`insert into task_exec_recode_` + strconv.Itoa(taskDto.Id) + ` (task_status, create_time) values 
	(1, now())`)
	var recordId int
	db.Raw(`select id from task_exec_recode_` + strconv.Itoa(taskDto.Id) + ` where task_status = 1`).Scan(&recordId)

	var stages []dto.TaskStages
	find = db.Raw("select * from task_stages where task_id = ? and status =1", taskDto.Id).Scan(&stages)
	if find.Error != nil {
		panic(find.Error)
	}
	for _, stage := range stages {
		db.Exec("insert into task_exec_stage_result_" + strconv.Itoa(taskDto.Id) + " (record_id, stage_type, stage_status) " +
			"values (?, ? , 0)", recordId, stage.StageType)
	}

	// 异步构建任务
	go task.RunTask(taskCode, taskDto.Id, recordId)

	response.Success(c, gin.H{"recordId": recordId}, "任务发起成功")
}

func GetTaskLogForWS(c *gin.Context) {
	recordId := c.Query("recordId")
	taskCode := c.Query("taskCode")
	taskId := c.Query("id")

	time.Sleep(1e9)	// wait job start

	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		panic(err)
	}
	defer ws.Close()

	reader, file := logFromFile(viper.GetString("taskLog") + taskCode + "/" + taskCode + "_" + recordId + ".log")
	defer file.Close()

	// buf := make([]byte, 64)
	db := common.GetDb()
	var taskStatus int
	for  {
		line, _, err := reader.ReadLine()
		// n, err := reader.Read(buf)
		if err == io.EOF {
			db.Raw("select task_status from task_exec_recode_" + taskId + " where id = ?", recordId).Scan(&taskStatus)
			if taskStatus != 1 {
				break
			}
			time.Sleep(1e9)
			//ws.Close()
			//fmt.Println("文件读完了")
			//break
		}
		if err != nil {
			log.Println(err)
		}
		if err == nil {
			err = ws.WriteMessage(websocket.TextMessage, line)
			if err != nil {
				panic(err)
			}
			time.Sleep(1e8)
		}
	}
}

func GetTaskLog(c *gin.Context) {
	recordId := c.Query("recordId")
	taskCode := c.Query("taskCode")
	LF := c.DefaultQuery("linefeed", "\n")

	reader, file := logFromFile(viper.GetString("taskLog") + taskCode + "/" + taskCode + "_" + recordId + ".log")
	defer file.Close()
	var buf []byte
	for {
		line, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		buf = append(buf, line...)
		buf = append(buf, []byte(LF)...)
	}

	response.Success(c, gin.H{"data": string(buf)}, "任务发起成功")
}

func logFromFile (filePath string) (*bufio.Reader, *os.File) {
	file, err := os.OpenFile(filePath, os.O_RDONLY, 0666)
	if err != nil {
		panic("没找到对应任务的历史记录，可能已清理！")
	}

	return bufio.NewReader(file), file
}

