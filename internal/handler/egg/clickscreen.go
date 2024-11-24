package egg

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

func ClickScreen(c *gin.Context) {
	ctx := c.Request.Context()
	roleId := contextx.FromRoleID(ctx)
	db := contextx.FromGormDB(ctx)

	req := new(schema.ClickScreenReq)
	if err := ginx.ParseJSON(c, req); err != nil {
		ginx.ResError(c, http.StatusOK, errors.NewResponseError(constant.ParametersInvalid, err))
		return
	}

	resp, err := logic.EggLogic.ClickScreen(ctx, db, roleId, req)
	if err != nil {
		ginx.ResError(c, http.StatusOK, err)
		return
	}
	ginx.ResData(c, constant.OK, resp)
}
