package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

//SetupRouter ...
func SetupRouter() *gin.Engine {
	conf, err := GetConfig("config.yaml")

	if err != nil {
		log.Fatal(err)
	}

	apiV := conf.APIVersion
	apps := conf.Apps

	r := gin.Default()
	v0 := r.Group("v" + apiV)

	for _, app := range apps {
		appR := v0.Group(app.PathPrefix)
		{
			appR.GET("/", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{
					"data": app.Data,
				})
			})
		}
	}

	return r
}
