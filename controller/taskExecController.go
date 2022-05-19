package controller

import (
	"com.csion/tasks/common"
	"com.csion/tasks/dto"
	"com.csion/tasks/response"
	"com.csion/tasks/task"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
	"strconv"
)

// 构建任务
func RunJob(c *gin.Context){
	taskCode := c.Query("taskCode")
	log.Debug("任务触发执行，任务编号：", taskCode)

	db := common.GetDb()
	var taskDto dto.Tasks
	log.Panic2("数据操作异常：", db.Where("task_code = ? and status =1", taskCode).Find(&taskDto).Error)

	// todo: 校验是否有可执行节点

	// 校验是否已经编排任务
	var n int64
	log.Panic2("数据操作异常：", common.GetDb().Model(&dto.TaskStages{}).Where("task_id = ? and status = 1", taskDto.Id).Count(&n).Error)
	if n == 0 {
		log.Panic1("该任务未编排任务节点信息，请先添加后再执行")
	}

	var recordId int
	// 根据taskCode查找task信息
	_ = common.GetDb().Transaction(func(db *gorm.DB) error {

		// 通过判断task的状态判断是否需要初始化任务执行记录表
		if taskDto.TaskStatus == 0 {
			log.Debug("初始化任务执行记录表，任务编号：", taskCode)
			log.Panic2("数据操作异常：", db.Exec(`create table task_exec_recode_`+strconv.Itoa(taskDto.Id)+` (
    id int(11) NOT NULL AUTO_INCREMENT COMMENT '主键',
	task_status tinyint(1) NOT NULL DEFAULT '1' COMMENT '任务执行状态：1：执行中，2：执行成功，3：执行失败',
    create_time datetime DEFAULT NULL COMMENT '创建时间',
    update_time datetime DEFAULT NULL COMMENT '更新时间',
    PRIMARY KEY (id)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8 COMMENT= '任务执行记录表'`).Error)
			log.Panic2("数据操作异常：", db.Exec(`create table task_exec_stage_result_`+strconv.Itoa(taskDto.Id)+` (
  id int(11) NOT NULL AUTO_INCREMENT COMMENT '主键',
  record_id int(11) NOT NULL COMMENT '执行记录标识',
  stage_type tinyint(1) NOT NULL COMMENT '节点类型',
  stage_status tinyint(1) NOT NULL DEFAULT '1' COMMENT '节点执行状态：1：执行中，2：执行成功，3：执行失败',
  create_time datetime DEFAULT NULL COMMENT '创建时间',
  update_time datetime DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='任务节点执行记录表'`).Error)
		}

		// 更新任务状态
		log.Panic2("数据操作异常：", db.Exec("update tasks set task_status = 1 where task_code = ?", taskCode).Error)

		// 向任务执行记录表中插入数据
		log.Panic2("数据操作异常：", db.Exec(`insert into task_exec_recode_`+strconv.Itoa(taskDto.Id)+` (task_status, create_time) values 
	(1, now())`).Error)

		log.Panic2("数据操作异常：", db.Raw(`select id from task_exec_recode_`+strconv.Itoa(taskDto.Id)+` where task_status = 1`).Scan(&recordId).Error)

		var stages []dto.TaskStages
		log.Panic2("数据操作异常：", db.Raw("select * from task_stages where task_id = ? and status =1", taskDto.Id).Scan(&stages).Error)
		for _, stage := range stages {
			log.Panic2("数据操作异常：", db.Exec("insert into task_exec_stage_result_"+strconv.Itoa(taskDto.Id)+" (record_id, stage_type, stage_status) "+
				"values (?, ? , 0)", recordId, stage.StageType).Error)
		}

		// 异步构建任务
		log.Debug("开启异步执行任务，任务编号：", taskCode)
		go task.PublishTask(taskCode, taskDto.Id, recordId)
		
		return nil
	})
	log.Debug("任务发起成功，任务编号：", taskCode)
	response.Success(c, gin.H{"recordId": recordId}, "任务发起成功")
}

// 获取任务执行记录，倒叙排序加分页
func GetTaskRecord(c *gin.Context)  {
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))

	taskId := c.Query("taskId")
	db := common.GetDb()

	var total int64
	log.Panic2("数据查询异常：", db.Raw("select count(1) from task_exec_recode_"+taskId).Scan(&total).Error)

	var recodes []dto.TaskExecRecode
	log.Panic2("数据查询异常：", db.Raw("select * from task_exec_recode_"+taskId+" order by id desc limit ?, ?", (page-1)*pageSize, pageSize).Scan(&recodes).Error)

	for n, recode := range recodes {
		var result []dto.TaskExecStageResult
		log.Panic2("数据查询异常：", db.Raw("select * from task_exec_stage_result_"+taskId+" where record_id = ? order by id asc ", recode.Id).Scan(&result).Error)
		var LastStageStatus int
		var LastStage = 1
		for i, stageResult := range result {
			result[i].StageTypeStr = getStageName(stageResult.StageType)
			if stageResult.StageStatus > 0 {
				LastStageStatus = stageResult.StageStatus
				LastStage = i
			}
		}
		recodes[n].StageResult = result
		recodes[n].LastStage = LastStage
		recodes[n].LastStageStatus = getStageStatus(LastStageStatus)
	}

	response.Success(c, gin.H{"data": recodes, "total": total}, "查询成功")
}

func getStageStatus(stageType int) string {
	switch stageType {
	case 1:
		return "finish"
	case 2:
		return "success"
	case 3:
		return "error"
	default:
		return "finish"
	}
}

// ws更新正在执行的任务状态
func UpdateTaskRecord(c *gin.Context)  {
	recordId := c.Query("recordId")
	taskId := c.Query("taskId")

	ch := task.GetTaskStageChan(taskId, recordId)

	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	log.Panic2("更新正在执行的任务状态异常：", err)
	defer ws.Close()
	defer task.RemoveChan(taskId, recordId)

	for {
		stage := <- ch
		err = ws.WriteMessage(websocket.TextMessage, []byte(strconv.Itoa(stage)))
		log.Panic2("更新正在执行的任务状态异常：", err)
		if stage <= 0 {
			break
		}
	}
}
