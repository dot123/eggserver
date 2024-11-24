package user

import (
	"eggServer/internal/constant"
	"eggServer/internal/contextx"
	"eggServer/internal/ginx"
	"eggServer/internal/logic"
	"eggServer/internal/models"
	"github.com/gin-gonic/gin"
	"net/http"
)

// @Tags     用户模块
// @Summary  自动加体力
// @Accept   application/json
// @Produce  application/json
// @Security ApiKeyAuth
// @Success  200 {object} ginx.ResponseData{} "成功结果"
// @Failure  200 {object} ginx.ResponseFail{} "失败结果"
// @Router   /autoaddvit [post]
func AutoAddVit(c *gin.Context) {
	ctx := c.Request.Context()
	roleId := contextx.FromRoleID(ctx)
	db := contextx.FromGormDB(ctx)

	role, err := models.RoleRepo.Get(ctx, db, roleId)
	if err != nil {
		ginx.ResError(c, http.StatusOK, err)
		return
	}

	resp, err := logic.UserLogic.AutoAddVit(ctx, db, roleId, role.LastAddVitTime)
	if err != nil {
		ginx.ResError(c, http.StatusOK, err)
		return
	}
	ginx.ResData(c, constant.OK, resp)
}
