package middleware

import (
	"com.csion/tasks/response"
	"fmt"
	"github.com/gin-gonic/gin"
)

// 统一异常处理中间件
func RecoveryMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer func() {
			if err := recover();err != nil {
				response.Fail(ctx, nil, fmt.Sprint(err))
			}
		}()
		ctx.Next()
	}
}
