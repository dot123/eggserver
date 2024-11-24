package user

import (
	"eggServer/internal/constant"
	"eggServer/internal/contextx"
	"eggServer/internal/ginx"
	"eggServer/internal/logic"
	"eggServer/internal/schema"
	"eggServer/pkg/errors"
	"github.com/gin-gonic/gin"
	"net/http"
)

// @Tags     用户模块
// @Summary  GM
// @Accept   application/json
// @Produce  application/json
// @Security ApiKeyAuth
// @Param    data body     schema.GMReq        true "请求参数"
// @Success  200  {object} ginx.ResponseData{} "成功结果"
// @Failure  200  {object} ginx.ResponseFail{} "失败结果"
// @Router   /gm [post]
func GM(c *gin.Context) {
	ctx := c.Request.Context()
	roleId := contextx.FromRoleID(ctx)
	db := contextx.FromGormDB(ctx)

	req := new(schema.GMReq)
	if err := ginx.ParseJSON(c, req); err != nil {
		ginx.ResError(c, http.StatusOK, errors.NewResponseError(constant.ParametersInvalid, err))
		return
	}

	resp, err := logic.UserLogic.GM(ctx, db, roleId, req.Cmd)
	if err != nil {
		ginx.ResError(c, http.StatusOK, err)
		return
	}
	ginx.ResData(c, constant.OK, resp)
}
