package passportreward

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

func Get(c *gin.Context) {
	ctx := c.Request.Context()
	db := contextx.FromGormDB(ctx)
	roleId := contextx.FromRoleID(ctx)

	req := new(schema.PassPortRewardGetReq)
	if err := ginx.ParseJSON(c, req); err != nil {
		ginx.ResError(c, http.StatusOK, errors.NewResponseError(constant.ParametersInvalid, err))
		return
	}

	resp, err := logic.PassPortLogic.Get(ctx, db, roleId, req.Id, req.Deluxe == 1)
	if err != nil {
		ginx.ResError(c, http.StatusOK, err)
		return
	}

	ginx.ResData(c, constant.OK, resp)
}