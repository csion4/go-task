package controller

import (
	"com.csion/tasks/common"
	"com.csion/tasks/dto"
	"com.csion/tasks/module"
	"com.csion/tasks/response"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"time"
)

func Register(c *gin.Context){
	// 绑定参数
	var user dto.Users
	err := c.Bind(&user)
	if err != nil {
		panic(err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	user.Password = string(hashedPassword)

	// 存表
	db := common.GetDb()
	user.CreateTime = time.Now()
	user.CreateUser = 1
	db = db.Create(&user)
	if db.Error != nil {
		panic(db.Error)
	}

	response.Success(c, nil, "注册成功")
}

func Login(c *gin.Context){
	var login module.Login
	errBind := c.Bind(&login)
	if errBind != nil {
		panic(errBind)
	}

	var user dto.Users
	db := common.GetDb()
	db = db.Where("name = ?", login.Name).Find(&user)
	if db.Error != nil {
		panic(db.Error)
	}
	if user.Id == 0 {
		response.Fail(c, nil, "用户名密码错误")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(login.Password));err != nil {
		response.Fail(c, nil, "用户名密码错误")
		return
	}

	// 密码正确发放token给前端
	token,errToken := common.ReleaseToken(user)
	if errToken != nil{
		response.Fail(c, nil, "生成token失败")
		return
	}

	response.Success(c, gin.H{"token": token, "userName": login.Name}, "登陆成功")
}