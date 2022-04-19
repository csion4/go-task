// +build swagger

package main

import (
	_ "com.csion/tasks/docs"	// 主动引入我们的docs，否则它会找默认的然后就报错了
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
)

// 配置swaggerHandler
func init() {
	swaggerHandler = ginSwagger.WrapHandler(swaggerFiles.Handler)
}
