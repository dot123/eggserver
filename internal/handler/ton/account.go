package ton

import (
	"eggServer/internal/contextx"
	"eggServer/internal/logic"
	"github.com/gin-gonic/gin"
	"github.com/tonkeeper/tonapi-go"
	"net/http"
)

func Account(c *gin.Context) {
	ctx := c.Request.Context()
	logger := contextx.FromLogger(ctx)

	accountId := c.Query("accountId")
	resp, err := logic.TonapiLogic.TonApi().GetAccount(ctx, tonapi.GetAccountParams{AccountID: accountId})
	if err != nil {
		logger.Errorln(err)
		c.JSON(http.StatusOK, gin.H{"error": err})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, resp)
	c.Abort()
}
