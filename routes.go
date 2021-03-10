package main

import (
	"github.com/gin-gonic/gin"
)

// initRouter initializes router and creates routes for each application
func initRouter() *gin.Engine {
	r := gin.Default()
	v := r.Group("v" + Project.APIVersion)

	for _, app := range Project.Apps {
		appR := v.Group(app.PathPrefix)
		for _, authNVariant := range app.AuthN {
			appR.POST(authNVariant.Path, authNHandler(&app, &authNVariant))
		}

		//appR.POST("/register", registerHandler(&app))
		//appR.POST("/login", authNHandler(&app))
	}

	return r
}
