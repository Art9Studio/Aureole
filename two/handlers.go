package two

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

//GetConstString ...
func GetConstString(c *gin.Context) {
	const s string = "Hello world!"

	c.JSON(http.StatusOK, gin.H{
		"data": s,
	})
}
