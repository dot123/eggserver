package shop

import (
	"eggServer/internal/constant"
	"eggServer/internal/contextx"
	"eggServer/internal/ginx"
	"eggServer/internal/logic"
	"github.com/gin-gonic/gin"
	"net/http"
)

func Refresh(c *gin.Context) {
	ctx := c.Request.Context()
	roleId := contextx.FromRoleID(ctx)
	db := contextx.FromGormDB(ctx)

	resp, err := logic.ShopLogic.Refresh(ctx, db, roleId, false)
	if err != nil {
		ginx.ResError(c, http.StatusOK, err)
		return
	}

	ginx.ResData(c, constant.OK, resp)
}
