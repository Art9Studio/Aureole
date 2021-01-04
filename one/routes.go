package one

import "github.com/gin-gonic/gin"

//RegisterRoutes ...
func RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/", GetConstString)
}
