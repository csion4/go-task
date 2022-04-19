package common

import (
	_ "com.csion/tasks/config" // 加载配置文件
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

// 初始化数据库连接
func init(){
	var err error
	mysqlUrl := viper.GetString("mysql.dsn")
	printSql := viper.GetString("gorm.printSql")

	var config gorm.Config
	if printSql == "true" {
		config = gorm.Config{Logger: logger.Default.LogMode(logger.Info)}
	} else {
		config = gorm.Config{}
	}

	db, err = gorm.Open(mysql.Open(mysqlUrl), &config)

	if err != nil {
		panic(err)
	}

	// 使用gorm创建表
	//e := db.AutoMigrate(&module.User{})
	//if e != nil {
	//	panic(e)
	//}
}

func GetDb() *gorm.DB {
	return db
}
