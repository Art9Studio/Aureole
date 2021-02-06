package main

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gouth/storage"
	"k8s.io/client-go/util/jsonpath"
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

			appR.POST("/register", func(c *gin.Context) {
				var user User
				err := c.ShouldBindWith(&user, binding.Form)

				_, _ = app.Session.InsertUser(app.Auth.UserColl, *storage.NewInsertUserData("1", "password"))
			})
		}
	}

	return r
}
