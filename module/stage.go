package module

type LayoutTask struct {
	Stages *[]Stage	`json:"stages"`
	TaskId int	`json:"taskId"`
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