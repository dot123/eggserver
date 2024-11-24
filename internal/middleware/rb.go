package middleware

import (
	"eggServer/internal/contextx"
	"eggServer/pkg/redisbackend"
	"github.com/gin-gonic/gin"
)

func RB(rd *redisbackend.RedisBackend) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		ctx = contextx.NewRB(ctx, rd)
		c.Request = c.Request.WithContext(ctx)

		// 处理请求
		c.Next()
	}
}
