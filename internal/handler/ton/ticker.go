package ton

import (
	"eggServer/internal/constant"
	"eggServer/internal/ginx"
	"eggServer/internal/logic"
	"eggServer/internal/schema"
	"eggServer/pkg/errors"
	"github.com/gin-gonic/gin"
	"net/http"
)

func Ticker(c *gin.Context) {
	ctx := c.Request.Context()

	req := new(schema.TickerReq)
	if err := ginx.ParseJSON(c, req); err != nil {
		ginx.ResError(c, http.StatusOK, errors.NewResponseError(constant.ParametersInvalid, err))
		return
	}

	last, err := logic.TonapiLogic.Ticker(ctx, req.InstId)
	if err != nil {
		ginx.ResError(c, http.StatusOK, err)
		return
	}

	resp := new(schema.TickerResp)
	resp.Last = last
	ginx.ResData(c, constant.OK, resp)
}
