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
// @Summary  登录
// @Accept   application/json
// @Produce  application/json
// @Security ApiKeyAuth
// @Param    data body     schema.LoginReq     true "请求参数"
// @Success  200  {object} ginx.ResponseData{} "成功结果"
// @Failure  200  {object} ginx.ResponseFail{} "失败结果"
// @Router   /login [post]
func Login(c *gin.Context) {
	ctx := c.Request.Context()
	db := contextx.FromGormDB(ctx)

	req := new(schema.LoginReq)
	if err := ginx.ParseJSON(c, req); err != nil {
		ginx.ResError(c, http.StatusBadRequest, errors.NewResponseError(constant.ParametersInvalid, err))
		return
	}

	token, err := logic.UserLogic.Login(ctx, db, req)
	if err != nil {
		ginx.ResError(c, http.StatusOK, err)
		return
	}

	// 登录成功，返回 token
	ginx.ResData(c, constant.OK, &schema.LoginResp{Token: token})
}
