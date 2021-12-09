package controller

import (
	"com.csion/tasks/common"
	"com.csion/tasks/dto"
	"com.csion/tasks/module"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

// 任务步骤编排
func LayoutTask(c *gin.Context){
	defer err(c)

	var stages module.LayoutTask
	err := c.Bind(&stages)
	if err != nil {
		panic(err)
	}
	// fmt.Println(stages)

	// 存表
	taskId := stages.TaskId
	s := *stages.Stages
	var taskEnvs []dto.TasksEnvs
	db := common.GetDb()

	for _, v := range s {
		var taskStage dto.TaskStages
		taskStage.TaskId = taskId
		taskStage.StageType = v.StageType
		taskStage.CreateTime = time.Now()
		taskStage.UpdateTime = time.Now()
		taskStage.CreateUser = 1

		create := db.Create(&taskStage)
		if create.Error != nil {
			panic(create.Error)
		}
		for k, e := range v.Envs {
			var env dto.TasksEnvs
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

	c.JSON(http.StatusOK, gin.H{
		"data": "保存成功",
	})
}

func LayoutInfo(c *gin.Context){

}
