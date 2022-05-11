package vo

type LayoutTask struct {
	Stages []Stage	`json:"stages" binding:"required"`
	TaskId int	`json:"taskId" binding:"required"`
}

type Stage struct {
	StageType int		`json:"stageType"`
	StageName string 	`json:"stageName"`
	Envs map[string]string	`json:"envs"`
}

type LayoutInfo struct {
	Id int
	TaskId int
	StageType int
	OrderBy int
	Param string
	Value string
}