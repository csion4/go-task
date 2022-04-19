package controller

import (
	"com.csion/tasks/common"
	"com.csion/tasks/dto"
	"com.csion/tasks/module"
	"com.csion/tasks/response"
	"com.csion/tasks/task"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"strconv"
	"time"
)

// 任务步骤编排
func LayoutTask(c *gin.Context){
	var stages module.LayoutTask
	err := c.ShouldBind(&stages)
	if err != nil {
		panic(err)
	}

	// todo：校验正在执行的任务，不可以修改节点信息

	taskId := stages.TaskId
	s := *stages.Stages
	var taskEnvs []dto.TaskEnvs
	db := common.GetDb()

	// 清除原来数据
	db.Exec("update task_stages set status = 0, update_time = now(), update_user = ? where task_id = ?", 1, taskId)

	// 存表
	for n, v := range s {
		var taskStage dto.TaskStages
		taskStage.TaskId = taskId
		taskStage.StageType = v.StageType
		taskStage.OrderBy = n
		taskStage.CreateTime = time.Now()
		taskStage.UpdateTime = time.Now()
		taskStage.CreateUser = 1

		create := db.Create(&taskStage)
		if create.Error != nil {
			panic(create.Error)
		}
		for k, e := range v.Envs {
			var env dto.TaskEnvs
			env.TaskId = taskId
			env.StageId = taskStage.Id
			env.Param = k
			env.Value = e
			env.CreateTime = time.Now()
			env.UpdateTime = time.Now()
			env.CreateUser = 1
			taskEnvs = append(taskEnvs, env)
		}
	}

	create := db.Create(&taskEnvs)
	if create.Error != nil {
		panic(create.Error)
	}

	response.Success(c, nil, "保存成功")
}

// 获取任务节点信息
func TaskLayoutInfo(c *gin.Context){
	taskId := c.Query("taskId")

	db := common.GetDb()
	var layoutInfo []module.LayoutInfo
	result := db.Raw(`select a.id, a.task_id, a.stage_type, a.order_by , b.param , b.value  from task_stages a, task_envs b
where a.task_id = ` + taskId + ` and a.status = 1 and a.id = b.stage_id and b.status = 1 order by a.order_by asc `).Scan(&layoutInfo)

	if result.Error != nil {
		panic(result.Error)
	}

	var temp = -1
	var stages []module.Stage
	var stage module.Stage
	var envs map[string]string
	for _, info := range layoutInfo {
		if temp != info.OrderBy {
			if temp != -1 {stages = append(stages, stage)}
			envs = make(map[string]string, 10)
			stageName := getStageName(info.StageType)
			stage = module.Stage{
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

func getStageName(StageStatus int) (string) {
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

func getStageStatus(stageType int) (string) {
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

// 获取任务执行记录，倒叙排序加分页
func GetTaskRecord(c *gin.Context)  {
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))

	taskId := c.Query("taskId")

	var total int64
	result := common.GetDb().Raw("select count(1) from task_exec_recode_"+taskId).Scan(&total)
	if result.Error != nil {
		panic(result.Error)
	}
	var recodes []dto.TaskExecRecode
		result = common.GetDb().Raw("select * from task_exec_recode_"+taskId+" order by id desc limit ?, ?", (page-1)*pageSize, pageSize).Scan(&recodes)
	if result.Error != nil {
		panic(result.Error)
	}

	for n, recode := range recodes {
		var result []dto.TaskExecStageResult
		scan := common.GetDb().Raw("select * from task_exec_stage_result_"+taskId+" where record_id = ? order by id asc ", recode.Id).Scan(&result)
		if scan.Error != nil {
			panic(scan.Error)
		}
		var LastStageStatus int
		var LastStage = 1
		for i, stageResult := range result {
			result[i].StageTypeStr = getStageName(stageResult.StageType)
			if (stageResult.StageStatus > 0){
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

// ws更新正在执行的任务状态
func UpdateTaskRecord(c *gin.Context)  {
	recordId := c.Query("recordId")
	taskId := c.Query("taskId")

	ch := task.GetTaskStageChan(taskId, recordId)

	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		panic(err)
	}
	defer ws.Close()
	defer task.RemoveChan(taskId, recordId)

	for ;; {
		stage := <- ch
		err = ws.WriteMessage(websocket.TextMessage, []byte(strconv.Itoa(stage)))
		if err != nil {
			panic(err)
		}
		if stage <= 0 {
			break
		}
	}
}
