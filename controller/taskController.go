package controller

import (
	"com.csion/tasks/common"
	"com.csion/tasks/dto"
	"com.csion/tasks/task"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"time"
)

// 异常处理
func err(c *gin.Context) {
	r := recover()
	if r != nil {
		log.Println("接口调用异常", r) // 这里可以优化一下日志打印
		c.JSON(http.StatusInternalServerError, gin.H{
			"data": "接口异常",
		})
	}
}

// hello
func Hello(c *gin.Context){
	c.JSON(http.StatusOK, gin.H{
		"data": "hello",
	})
}

// 添加任务
func AddJob(c *gin.Context){
	defer err(c)

	// 绑定参数
	var tasks dto.Tasks
	err := c.Bind(&tasks)
	if err != nil {
		panic(err)
	}
	tasks.CreateTime = time.Now()
	tasks.CreateUser = 1
	// fmt.Println("tasks实体参数：", tasks)

	// 存表
	db := common.GetDb()
	result := db.Create(&tasks)
	if result.Error != nil {
		panic(result.Error)
	}

	c.JSON(http.StatusOK, gin.H{
		"data": "添加成功",
	})
}

// 删除任务
func DelJob(c *gin.Context){
	defer err(c)

	taskId := c.Query("taskId")

	db := common.GetDb()
	result := db.Exec("update tasks set status = 0 where id = ?", taskId)
	if result.Error != nil {
		panic(result.Error)
	}

	c.JSON(http.StatusOK, gin.H{
		"data": "删除成功",
	})
}

//  查询任务
func GetJobs(c *gin.Context){
	defer err(c)

	var tasks []dto.Tasks
	db := common.GetDb()
	result := db.Where("status = ?", 1).Find(&tasks)
	if result.Error != nil {
		panic(result.Error)
	}

	c.JSON(http.StatusOK, gin.H{
		"data": tasks,
	})
}


// 构建任务
func RunJob(c *gin.Context){
	defer err(c)

	taskCode := c.Query("taskCode")
	go task.RunTask(taskCode)

	c.JSON(http.StatusOK, gin.H{
		"data": "任务发起成功",
	})
}
