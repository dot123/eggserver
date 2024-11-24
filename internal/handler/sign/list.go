package sign

import (
	"eggServer/internal/constant"
	"eggServer/internal/contextx"
	"eggServer/internal/ginx"
	"eggServer/internal/logic"
	"github.com/gin-gonic/gin"
	"net/http"
)

func Data(c *gin.Context) {
	ctx := c.Request.Context()
	db := contextx.FromGormDB(ctx)
	roleId := contextx.FromRoleID(ctx)

	resp, err := logic.SignLogic.Data(ctx, db, roleId)
	if err != nil {
		ginx.ResError(c, http.StatusOK, err)
		return
	}

	ginx.ResData(c, constant.OK, resp)
}
