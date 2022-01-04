package dto

import "time"

type Users struct {
	Id int	`json:"id"`
	Name string	`json:"name"`
	Password string `json:"password"`
	CreateTime time.Time `json:"createTime"`
	UpdateTime time.Time `gorm:"default:null" json:"updateTime"`
	CreateUser int	`json:"createUser"`
	UpdateUser int	`gorm:"default:null" json:"updateUser"`
}

