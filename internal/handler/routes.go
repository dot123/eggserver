package handler

import (
	"eggServer/internal/config"
	"eggServer/internal/handler/battle"
	"eggServer/internal/handler/egg"
	"eggServer/internal/handler/guide"
	"eggServer/internal/handler/item"
	"eggServer/internal/handler/leaderboard"
	"eggServer/internal/handler/passportreward"
	"eggServer/internal/handler/platform"
	"eggServer/internal/handler/shop"
	"eggServer/internal/handler/sign"
	"eggServer/internal/handler/task"
	"eggServer/internal/handler/ton"
	"eggServer/internal/handler/user"
	"eggServer/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// RegisterHandlers 注册路由
func RegisterHandlers(app *gin.Engine, client redis.UniversalClient, l *logrus.Entry) {
	app.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	g := app.Group("/api")
	v1 := g.Group("/v1")
	v1.Use(middleware.Logger(l))

	v1.POST("/login", user.Login)
	v1.POST("/transaction", ton.Transaction)
	v1.POST("/account", ton.Account)
	v1.POST("/ticker", ton.Ticker)
	v1.POST("/platformurl", platform.Url)

	v1.Use(middleware.Auth())
	v1.Use(middleware.RateLimiter(client))

	if config.C.GM {
		v1.POST("/gm", user.GM)
	}

	v1.POST("/enter", user.Enter)
	v1.POST("/daily", user.Daily)
	v1.POST("/autoaddvit", user.AutoAddVit)
	v1.POST("/syncdata", user.SyncData)

	v1.POST("/autolayegg", egg.AutoLayEgg)
	v1.POST("/clickscreen", egg.ClickScreen)
	v1.POST("/eggopen", egg.Open)

	v1.POST("/taskrewards", task.Rewards)
	v1.POST("/taskdata", task.Data)
	v1.POST("/taskgoto", task.Goto)
	v1.POST("/tasktonaccount", task.TonAccount)

	v1.POST("/shopdata", shop.Data)
	v1.POST("/shoprefresh", shop.Refresh)
	v1.POST("/shopbuy", shop.Buy)
	v1.POST("/shopshare", shop.Share)
	v1.POST("/createorder", shop.CreateOrder)
	v1.POST("/delivery", shop.Delivery)

	v1.POST("/battlematch", battle.Match)
	v1.POST("/battlematchstate", battle.MatchState)
	v1.POST("/battleleave", battle.Leave)
	v1.POST("/battlebet", battle.Bet)
	v1.POST("/battleroundresult", battle.RoundResult)
	v1.POST("/battlesettlement", battle.Settlement)
	v1.POST("/battleexit", battle.Exit)
	v1.POST("/battlesyncscore", battle.SyncScore)

	v1.POST("/itemuse", item.Use)
	v1.POST("/leaderboarddata", leaderboard.Data)

	v1.POST("/guidestep", guide.Step)

	v1.POST("/signdata", sign.Data)
	v1.POST("/signin", sign.In)

	v1.POST("/passportrewarddata", passportreward.Data)
	v1.POST("/passportrewardget", passportreward.Get)
}
