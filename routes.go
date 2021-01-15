package main

import (
	"github.com/gin-gonic/gin"
)

// initRouter initializes router and creates routes for each application
func initRouter() *gin.Engine {
	r := gin.Default()
	v := r.Group("v" + conf.APIVersion)

	for _, app := range conf.Apps {
		appR := v.Group(app.PathPrefix)
		{
			appR.GET("/", func(c *gin.Context) {
			})
		}
	}

	return r
}
