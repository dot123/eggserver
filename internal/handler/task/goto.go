package task

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

func Goto(c *gin.Context) {
	ctx := c.Request.Context()
	db := contextx.FromGormDB(ctx)
	roleId := contextx.FromRoleID(ctx)
	userId := contextx.FromUserID(ctx)
	req := new(schema.TaskGotoReq)
	if err := ginx.ParseJSON(c, req); err != nil {
		ginx.ResError(c, http.StatusOK, errors.NewResponseError(constant.ParametersInvalid, err))
		return
	}

	resp, err := logic.TaskLogic.RecordTaskProgressWithGoto(ctx, db, roleId, userId, req.UserUid, req.TaskSubId, 1)
	if err != nil {
		ginx.ResError(c, http.StatusOK, err)
		return
	}

	ginx.ResData(c, constant.OK, resp)
}
