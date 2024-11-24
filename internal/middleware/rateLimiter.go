package middleware

import (
	"eggServer/internal/config"
	"eggServer/internal/constant"
	"eggServer/internal/contextx"
	"eggServer/internal/ginx"
	"eggServer/pkg/errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
	"net/http"
)

func RateLimiter(client redis.UniversalClient) gin.HandlerFunc {
	cfg := config.C.RateLimiter
	if !cfg.Enable {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	limiter := redis_rate.NewLimiter(client)

	return func(c *gin.Context) {
		userID := contextx.FromUserID(c.Request.Context())
		if userID != 0 {
			res, err := limiter.Allow(c.Request.Context(), fmt.Sprintf("%d", userID), redis_rate.PerMinute(cfg.Count))
			if err != nil {
				ginx.ResError(c, http.StatusTooManyRequests, errors.New("limiter error"))
				return
			}

			if res.Allowed == 0 {
				ginx.ResError(c, http.StatusTooManyRequests, errors.NewResponseError(constant.TooManyRequests, errors.New("request Too Frequent")))
				return
			}
		}

		c.Next()
	}
}
