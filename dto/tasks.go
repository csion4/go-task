package dto

import (
	"com.csion/tasks/common"
	"time"
)

type Tasks struct {
	Id int			`json:"id"`
	Title string	`json:"title" binding:"required"`
	TaskCode string `json:"taskCode" binding:"required"`
	TaskStatus int
	CreateTime time.Time
	UpdateTime time.Time `gorm:"default:null"`
	CreateUser int
	UpdateUser int	`gorm:"default:null"`
}

func (t *Tasks) FindIdFromTaskCode(taskCode string) (id int) {
	log.Panic2("数据操作异常", common.GetDb().Model(&Tasks{}).Select("id").Where("task_code = ?", taskCode).Find(&id).Error)
	return
}

type TaskEnvs struct {
	Id int
	TaskId int
	StageId int
	Param string
	Value string
	CreateTime time.Time
	UpdateTime time.Time
	CreateUser int
	UpdateUser int
}

type TaskStages struct {
	Id int
	TaskId int
	StageType int
	OrderBy int
	CreateTime time.Time
	UpdateTime time.Time
	CreateUser int
	UpdateUser int
}

type TaskExecRecode struct {
	Id int	`json:"id"`
	TaskStatus int `json:"taskStatus"`
	CreateTime time.Time	`json:"createTime"`
	UpdateTime time.Time    `json:"updateTime"`

	StageResult []TaskExecStageResult `gorm:"-" json:"stageResult"`
	LastStage int `gorm:"-" json:"lastStage"`
	LastStageStatus string `gorm:"-" json:"lastStageStatus"`
}

type TaskExecStageResult struct {
	Id int		`json:"id"`
	RecordId int		`json:"recordId"`
	StageType int		`json:"stageType"`
	StageStatus int		`json:"stageStatus"`
	CreateTime time.Time		`json:"createTime"`
	UpdateTime time.Time		`json:"updateTime"`

	StageTypeStr string `json:"stageTypeStr" gorm:"-"`
}
