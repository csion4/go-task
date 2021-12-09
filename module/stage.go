package module

type LayoutTask struct {
	Stages *[]stage	`json:"stages"`
	TaskId int	`json:"taskId"`
}

type stage struct {
	StageType int		`json:"stageType"`
	Envs map[string]string	`json:"envs"`
}
