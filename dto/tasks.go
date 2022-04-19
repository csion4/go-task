package dto

import "time"

type Tasks struct {
	Id int			`json:"id"`
	Title string	`json:"title"`
	TaskCode string `json:"taskCode"`
	TaskStatus int
	CreateTime time.Time
	UpdateTime time.Time `gorm:"default:null"`
	CreateUser int
	UpdateUser int	`gorm:"default:null"`
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
