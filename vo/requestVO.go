package vo

// 请求的vo

// 注册请求
type RegisterReq struct {
	Name string `json:"name" binding:"required,max=10"`
	PassWord string `json:"passWord" binding:"required,max=20"`
}

// 登录请求
type LoginReq struct {
	Name string `json:"name" binding:"required"`
	Password string `json:"password" binding:"required"`
}
