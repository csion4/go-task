package dto

import "time"

type TasksEnvs struct {
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
