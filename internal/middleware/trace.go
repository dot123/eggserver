package middleware

import (
	"eggServer/internal/contextx"
	"fmt"
	"github.com/gin-gonic/gin"
	"os"
	"sync/atomic"
	"time"
)

var (
	incrNum uint64
	pid     = os.Getpid()
)

func NewTraceID() string {
	return fmt.Sprintf("%d-%s-%d", pid, time.Now().Format("2006.01.02.15.04.05.999"), atomic.AddUint64(&incrNum, 1))
}

func Trace() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader("X-Request-Id")
		if traceID == "" {
			traceID = NewTraceID()
		}

		ctx := contextx.NewTraceID(c.Request.Context(), traceID)
		c.Request = c.Request.WithContext(ctx)
		c.Writer.Header().Set("X-Trace-Id", traceID)

		c.Next()
	}
}
