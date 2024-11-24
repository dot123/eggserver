package middleware

import (
	"eggServer/internal/contextx"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const (
	TraceIDKey = "traceId"
	UserIDKey  = "userId"
	RoleIDKey  = "roleId"
)

func Logger(l *logrus.Entry) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		fields := logrus.Fields{}

		if v, ok := contextx.FromTraceID(ctx); ok {
			fields[TraceIDKey] = v
		}

		if v := contextx.FromUserID(ctx); v != 0 {
			fields[UserIDKey] = v
		}

		if v := contextx.FromRoleID(ctx); v != 0 {
			fields[RoleIDKey] = v
		}

		l = l.WithContext(ctx).WithFields(fields)
		ctx = contextx.NewLogger(ctx, l)
		c.Request = c.Request.WithContext(ctx)

		// 处理请求
		c.Next()
	}
}
