package dto

import "time"

type TaskStages struct {
	Id int
	TaskId int
	StageType int
	CreateTime time.Time
	UpdateTime time.Time
	CreateUser int
	UpdateUser int
}
