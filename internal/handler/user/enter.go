package user

import (
	"eggServer/internal/constant"
	"eggServer/internal/contextx"
	"eggServer/internal/ginx"
	"eggServer/internal/logic"
	"github.com/gin-gonic/gin"
	"net/http"
)

// @Tags     用户模块
// @Summary  处理用户进入游戏请求
// @Accept   application/json
// @Produce  application/json
// @Security ApiKeyAuth
// @Success  200 {object} ginx.ResponseData{} "成功结果"
// @Failure  200 {object} ginx.ResponseFail{} "失败结果"
// @Router   /enter [post]
func Enter(c *gin.Context) {
	ctx := c.Request.Context()
	db := contextx.FromGormDB(ctx)
	userId := contextx.FromUserID(ctx)

	resp, err := logic.UserLogic.Enter(ctx, db, userId)
	if err != nil {
		ginx.ResError(c, http.StatusOK, err)
		return
	}

	ginx.ResData(c, constant.OK, resp)
}
