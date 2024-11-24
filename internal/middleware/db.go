package middleware

import (
	"eggServer/internal/contextx"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func DB(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		ctx = contextx.NewGormDB(ctx, db)
		c.Request = c.Request.WithContext(ctx)

		// 处理请求
		c.Next()
	}
}
