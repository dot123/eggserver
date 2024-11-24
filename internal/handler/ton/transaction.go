package ton

import (
	"eggServer/internal/contextx"
	"eggServer/internal/logic"
	"github.com/gin-gonic/gin"
	"github.com/tonkeeper/tonapi-go"
	"net/http"
)

func Transaction(c *gin.Context) {
	ctx := c.Request.Context()
	logger := contextx.FromLogger(ctx)

	transactionId := c.Query("transactionId")
	resp, err := logic.TonapiLogic.TonApi().GetBlockchainTransaction(ctx, tonapi.GetBlockchainTransactionParams{TransactionID: transactionId})
	if err != nil {
		logger.Errorln(err)
		c.JSON(http.StatusOK, gin.H{"error": err})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, resp)
	c.Abort()
}
