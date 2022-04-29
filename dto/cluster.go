package dto

import "time"

type WorkerNode struct {
	Id int
	Ip string
	Port int
	UserName string
	Password string
	NodeStatus int
	CreateTime time.Time
	UpdateTime time.Time `gorm:"default:null"`
	CreateUser int
	UpdateUser int	`gorm:"default:null"`


}
