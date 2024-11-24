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

func TonAccount(c *gin.Context) {
	ctx := c.Request.Context()
	db := contextx.FromGormDB(ctx)
	roleId := contextx.FromRoleID(ctx)

	req := new(schema.TaskTonAccountReq)
	if err := ginx.ParseJSON(c, req); err != nil {
		ginx.ResError(c, http.StatusOK, errors.NewResponseError(constant.ParametersInvalid, err))
		return
	}

	err := logic.TaskLogic.TonAccount(ctx, db, roleId, req)
	if err != nil {
		ginx.ResError(c, http.StatusOK, err)
		return
	}

	ginx.ResOk(c)
}
