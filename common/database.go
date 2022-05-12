package common

import (
	_ "com.csion/tasks/config" // 加载配置文件
	"com.csion/tasks/tLog"
	tLogAdapter "com.csion/tasks/tLog/tLogApapter"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"time"
)

var db *gorm.DB

var log = tLog.GetTLog()
// 初始化数据库连接
func init(){
	log.Debug("开启建立数据库连接...")
	mysqlUrl := viper.GetString("mysql.dsn")
	printSql := viper.GetString("gorm.printSql")

	var LogLevel logger.LogLevel
	if printSql == "true" {
		LogLevel = logger.Info
	} else {
		LogLevel = logger.Error
	}

	var err error
	db, err = gorm.Open(mysql.Open(mysqlUrl),
		&gorm.Config{Logger: logger.New(
			&tLogAdapter.TLogGormAdapter{}, // 指定输出writer
			logger.Config{										// 增加配置
				SlowThreshold: 1 * time.Microsecond,			// 配置慢sql耗时标准，默认 200 * time.Millisecond
				LogLevel:      LogLevel,						// 打开Warn级别的日志，其实如果我们不需要修改其他配置比如SlowThreshold可以直接设置当前输出日志级别Warn即可
		}),
	})

	log.Panic2("GORM初始化异常：", err)
	log.Debug("数据库连接完成！")
}

func GetDb() *gorm.DB {
	return db
}