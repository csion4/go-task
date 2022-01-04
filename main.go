package main

import (
	"com.csion/tasks/common"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r = Route(r)
	common.GetDb()

	panic(r.Run(":9044"))

}
