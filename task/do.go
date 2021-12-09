package task

import (
	"com.csion/tasks/common"
	"com.csion/tasks/dto"
	"log"
)

const baseDir = "E:/tasks/"

// 构建任务
func RunTask(taskCode string){
	var stage []dto.TaskStages
	db := common.GetDb()

	find := db.Exec("select b.* from tasks a, task_stages b where a.task_code = ? and a.id = b.task_id and b.status = 1", taskCode).Find(&stage)
	if find.Error != nil {
		log.Fatal(find.Error)
	}

	for _, value := range stage {
		env := getEnv(value.Id)
		switch value.StageType {
		case 1:
			Git(env["gitUrl"], env["branch"], baseDir + taskCode)
			break
		case 2:
			// ExecScript()
			break
		case 3:
			// HttpInvoke()
			break
		default:
			log.Println("unSupport stage type")
		}
	}
}


func getEnv(stageId int) (env map[string]string) {

	var tasksEnv []dto.TasksEnvs
	db := common.GetDb()
	find := db.Where("stage_id = ? and status = 1", stageId).Find(&tasksEnv)
	if find.Error != nil {
		panic(find.Error)
	}

	env = make(map[string]string, len(tasksEnv))
	for _, v := range tasksEnv {
		env[v.Param] = v.Value
	}
	return env
}


