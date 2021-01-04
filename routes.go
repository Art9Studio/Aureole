package main

import (
	"./one"
	"./two"
	"github.com/gin-gonic/gin"
)

//SetupRouter ...
func SetupRouter() *gin.Engine {
	r := gin.Default()
	v := r.Group("/v0.1")

	one.RegisterRoutes(v.Group("/one"))
	two.RegisterRoutes(v.Group("/two"))

	return r
}
