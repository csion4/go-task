package controller

import (
	"com.csion/tasks/common"
	"com.csion/tasks/dto"
	"com.csion/tasks/response"
	"com.csion/tasks/tLog"
	"com.csion/tasks/vo"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"time"
)

var log = tLog.GetTLog()

// @Summary 用户注册
// @Description 用户注册
// @Tags UserController
// @Accept json
// @Produce json
// @Param param body vo.RegisterReq true "注册信息"
// @Success 200
// @Failure 400
// @Router /register [post]
func Register(c *gin.Context){
	// 绑定参数
	var req vo.RegisterReq
	log.Panic2("入参异常：", c.ShouldBind(&req))

	var user = &dto.Users{Name: req.Name, Password: req.PassWord}

	// 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	log.Panic2("加密异常：", err)

	user.Password = string(hashedPassword)

	// 存表
	db := common.GetDb()
	user.CreateTime = time.Now()
	user.CreateUser = 1
	db = db.Create(&user)
	log.Panic2("数据插入异常：", db.Error)

	response.Success(c, nil, "注册成功")
}

// @Summary 用户登录
// @Description 用户登录
// @Tags UserController
// @Accept mpfd
// @Param user formData string true "用户名"
// @Param password formData string true "密码"
// @Success 200
// @Failure 400
// @Router /login [post]
func Login(c *gin.Context){
	var login vo.LoginReq
	log.Panic2("入参异常：", c.ShouldBind(&login))

	var user dto.Users
	db := common.GetDb()
	db = db.Where("name = ?", login.Name).Find(&user)
	log.Panic2("用户查询异常：", db.Error)

	if user.Id == 0 {
		log.Panic1("用户名不存在")
	}

	log.Panic2("用户名密码错误", bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(login.Password)))

	// 密码正确发放token给前端
	token,errToken := common.ReleaseToken(user.Id)
	if errToken != nil{
		log.Panic2("生成token失败：", errToken)
	}

	log.Info("用户", login.Name, "登录")
	response.Success(c, gin.H{"token": token, "userName": login.Name}, "登陆成功")
}