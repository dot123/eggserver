package platform

import (
	"eggServer/pkg/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"net/http"
)

func Url(c *gin.Context) {
	p := c.Query("p")
	num := cast.ToUint64(p)
	if num > 1000 {
		c.JSON(http.StatusOK, gin.H{"error": "Invalid parameters"})
		c.Abort()
		return
	}

	url := fmt.Sprintf("https://t.me/eggroyale_bot/app?startapp=%s", utils.Encrypt(fmt.Sprintf("%d", num)))

	c.JSON(http.StatusOK, url)
	c.Abort()
}
