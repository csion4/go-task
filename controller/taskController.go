package controller

import (
	"bufio"
	"com.csion/tasks/common"
	"com.csion/tasks/config"
	"com.csion/tasks/dto"
	"com.csion/tasks/response"
	"com.csion/tasks/vo"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
	"gorm.io/gorm"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

var clusterTask = viper.GetString("task.cluster")

// 配置webSocket参数
var upgrader = websocket.Upgrader{
	ReadBufferSize: 10,
	WriteBufferSize: 512,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// 添加任务
func AddTask(c *gin.Context){
	// 绑定参数
	var tasks dto.Tasks
	log.Panic2("入参异常：", c.ShouldBind(&tasks))

	// 存表
	var r string
	db := common.GetDb()
	if tasks.Id == 0 {
		tasks.CreateTime = time.Now()
		tasks.CreateUser = c.GetInt("userId")
		db = db.Create(&tasks)
		r = "添加成功"
	} else {
		tasks.UpdateTime = time.Now()
		tasks.UpdateUser = c.GetInt("userId")
		db = db.Model(&tasks).Updates(&tasks)
		r = "修改成功"
	}
	log.Panic2("数据操作异常：", db.Error)

	response.Success(c, gin.H{"data": tasks}, r)
}

// 删除任务
func DelTask(c *gin.Context){
	taskId := c.Query("taskId")

	log.Panic2("数据操作异常：", common.GetDb().Exec("update tasks set status = 0 where id = ?", taskId).Error)

	response.Success(c, nil, "删除成功")
}

//  查询任务
func GetTasks(c *gin.Context){
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))

	var tasks []dto.Tasks
	var total int64
	log.Panic2("数据操作异常：", common.GetDb().Model(&dto.Tasks{}).Where("status = ?", 1).Count(&total).Error)
	log.Panic2("数据操作异常：", common.GetDb().Order("id desc").Limit(pageSize).Offset((page - 1) * pageSize).Where("status = ?", 1).Find(&tasks).Error)

	response.Success(c, gin.H{"data": tasks, "total": total}, "查询成功")
}

// 任务步骤编排
func LayoutTask(c *gin.Context){
	var stages vo.LayoutTask
	log.Panic2("入参异常：", c.ShouldBind(&stages))

	taskId := stages.TaskId
	db := common.GetDb()
	// 校验正在执行的任务，不可以修改节点信息
	var taskStatus int
	db.Table("tasks").Select("task_status").Where("id = ?", taskId).Find(&taskStatus)
	if taskStatus == 1 {
		log.Panic1("该任务真在执行，请勿修改！")
	}

	var taskEnvs []dto.TaskEnvs
	_ = db.Transaction(func(tx *gorm.DB) error {
		// 清除原来数据
		tx.Exec("update task_stages set status = 0, update_time = now(), update_user = ? where task_id = ?", 1, taskId)

		// 存表
		for n, v := range stages.Stages {
			var taskStage dto.TaskStages
			taskStage.TaskId = taskId
			taskStage.StageType = v.StageType
			taskStage.OrderBy = n
			taskStage.CreateTime = time.Now()
			taskStage.UpdateTime = time.Now()
			taskStage.CreateUser = c.GetInt("userId")
			log.Panic2("数据操作异常：", tx.Create(&taskStage).Error)

			for k, e := range v.Envs {
				var env dto.TaskEnvs
				env.TaskId = taskId
				env.StageId = taskStage.Id
				env.Param = k
				env.Value = e
				env.CreateTime = time.Now()
				env.UpdateTime = time.Now()
				env.CreateUser = c.GetInt("userId")
				taskEnvs = append(taskEnvs, env)
			}
		}
		log.Panic2("数据操作异常：", tx.Create(&taskEnvs).Error)
		return nil
	})

	response.Success(c, nil, "保存成功")
}


// 获取任务节点信息
func TaskLayoutInfo(c *gin.Context){
	taskId := c.Query("taskId")

	db := common.GetDb()
	var layoutInfo []vo.LayoutInfo
	log.Panic2("数据操作异常：", db.Raw(`select a.id, a.task_id, a.stage_type, a.order_by , b.param , b.value  from task_stages a, task_envs b
where a.task_id = ? and a.status = 1 and a.id = b.stage_id and b.status = 1 order by a.order_by asc `, taskId).Scan(&layoutInfo).Error)

	var temp = -1
	var stages []vo.Stage
	var stage vo.Stage
	var envs map[string]string
	for _, info := range layoutInfo {
		if temp != info.OrderBy {
			if temp != -1 {stages = append(stages, stage)}
			envs = make(map[string]string, 10)
			stageName := getStageName(info.StageType)
			stage = vo.Stage{
				StageType: info.StageType,
				StageName: stageName,
				Envs: envs,
			}
			temp = info.OrderBy
		}
		envs[info.Param] = info.Value
	}
	stages = append(stages, stage)

	response.Success(c, gin.H{"data": stages}, "查询成功")
}

func getStageName(StageStatus int) string {
	switch StageStatus {
	case 1:
		return "更新Git仓库代码"
	case 2:
		return "脚本执行"
	case 3:
		return "Http/Https服务调用"
	case 4:
		return "远程ssh访问"
	default:
		return "未知步骤"
	}
}



func GetTaskLogForWS(c *gin.Context) {
	recordId := c.Query("recordId")
	taskCode := c.Query("taskCode")
	taskId := c.Query("id")

	time.Sleep(2e9)	// wait job start

	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	log.Panic2("WS获取执行日志异常：", err)
	defer ws.Close()

	_, logFilePath := config.GetLogFilePath(taskCode, recordId)
	reader, file := logFromFile(logFilePath)
	defer file.Close()

	db := common.GetDb()
	var taskStatus int
	for  {
		line, _, err := reader.ReadLine()
		if err == io.EOF {
			db.Raw("select task_status from task_exec_recode_" + taskId + " where id = ?", recordId).Scan(&taskStatus)
			if taskStatus != 1 {
				break
			}
			time.Sleep(1e9)
		}
		if err != nil {
			log.Panic2("WS获取执行日志异常：", err)
		}
		if err == nil {
			err = ws.WriteMessage(websocket.TextMessage, line)
			log.Panic2("WS获取执行日志异常：", err)
			time.Sleep(1e8)
		}
	}
}

func GetTaskLog(c *gin.Context) {
	recordId := c.Query("recordId")
	taskCode := c.Query("taskCode")
	LF := c.DefaultQuery("linefeed", "\n")

	_, logFilePath := config.GetLogFilePath(taskCode, recordId)
	reader, file := logFromFile(logFilePath)
	defer file.Close()
	var buf []byte
	for {
		line, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Panic2("获取执行日志异常：", err)
		}
		buf = append(buf, line...)
		buf = append(buf, []byte(LF)...)
	}

	response.Success(c, gin.H{"data": string(buf)}, "任务发起成功")
}

func logFromFile (filePath string) (*bufio.Reader, *os.File) {
	file, err := os.OpenFile(filePath, os.O_RDONLY, 0666)
	if err != nil {
		log.Panic1("没找到对应任务的历史记录，可能已清理！")
	}
	return bufio.NewReader(file), file
}

