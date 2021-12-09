package dto

import "time"

type Tasks struct {
	Id int			`json:"id"`
	Title string	`json:"title"`
	CreateTime time.Time `json:"createTime"`
	UpdateTime time.Time `gorm:"default:null" json:"updateTime"`
	CreateUser int	`json:"createUser"`
	UpdateUser int	`gorm:"default:null" json:"updateUser"`
}
