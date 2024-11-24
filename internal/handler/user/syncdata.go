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
// @Summary  同步数据
// @Accept   application/json
// @Produce  application/json
// @Security ApiKeyAuth
// @Success  200 {object} ginx.ResponseData{} "成功结果"
// @Failure  200 {object} ginx.ResponseFail{} "失败结果"
// @Router   /syncdata [post]
func SyncData(c *gin.Context) {
	ctx := c.Request.Context()
	roleId := contextx.FromRoleID(ctx)
	db := contextx.FromGormDB(ctx)

	resp, err := logic.UserLogic.SyncData(ctx, db, roleId)
	if err != nil {
		ginx.ResError(c, http.StatusOK, err)
		return
	}
	ginx.ResData(c, constant.OK, resp)
}
