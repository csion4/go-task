package common

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB
// 初始化数据库连接，获取数据库实例
func initDB(){
	var err error
	db, err = gorm.Open(mysql.Open("root:123456@tcp(127.0.0.1:3306)/task?"+
		"charset=utf8&parseTime=True&loc=Asia%2FShanghai"),
		&gorm.Config{})

	if err != nil {
		panic(err)
	}

	// 使用gorm创建表
	//e := db.AutoMigrate(&module.User{})
	//if e != nil {
	//	panic(e)
	//}
}

// 需要单例吗？应该是启动时就初始化了，可以暂时不用单例
func GetDb() *gorm.DB {
	if db == nil {
		initDB()
	}
	return db
}
