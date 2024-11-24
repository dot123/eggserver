package user

import (
	"eggServer/internal/constant"
	"eggServer/internal/contextx"
	"eggServer/internal/ginx"
	"eggServer/internal/logic"
	"eggServer/internal/models"
	"eggServer/pkg/utils"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

// @Tags     用户模块
// @Summary  每日操作
// @Accept   application/json
// @Produce  application/json
// @Security ApiKeyAuth
// @Success  200 {object} ginx.ResponseData{} "成功结果"
// @Failure  200 {object} ginx.ResponseFail{} "失败结果"
// @Router   /daily [post]
func Daily(c *gin.Context) {
	ctx := c.Request.Context()
	roleId := contextx.FromRoleID(ctx)
	db := contextx.FromGormDB(ctx)

	role, err := models.RoleRepo.Get(ctx, db, roleId)
	if err != nil {
		ginx.ResError(c, http.StatusOK, err)
		return
	}

	// 每天操作
	now := time.Now().Unix()
	if role.LastDailyTime == 0 || utils.IsDifferentDays(time.Unix(role.LastDailyTime, 0), time.Unix(now, 0), "Asia/Shanghai") {
		resp, err := logic.UserLogic.Daily(ctx, db, role)
		if err != nil {
			ginx.ResError(c, http.StatusOK, err)
			return
		}

		if err := models.RoleRepo.Updates(ctx, db, role.ID, map[string]interface{}{"lastDailyTime": role.LastDailyTime,
			"lastLogin": role.LastLogin}); err != nil {
			ginx.ResError(c, http.StatusOK, err)
			return
		}
		ginx.ResData(c, constant.OK, resp)
		return
	}

	ginx.ResOk(c)
}
