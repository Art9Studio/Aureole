package main

import (
	"github.com/gin-gonic/gin"
	"gouth/storage"
	"net/http"
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
				var registerData interface{}
				err := c.BindJSON(&registerData)
				if err != nil {
					c.Error(err)
					c.AbortWithStatusJSON(http.StatusInternalServerError, err)
				}

				var registerConfig = app.Auth.Register
				mapKeys := registerConfig.Fields

				userUnique, err := GetJSONPath(mapKeys["user_unique"], registerData)
				if err != nil {
					c.Error(err)
					c.AbortWithStatusJSON(
						http.StatusInternalServerError,
						map[string]string{"code": "user_unique didn't passed"},
					)
					return
				}

				userConfirm, err := GetJSONPath(mapKeys["user_confirm"], registerData)
				if err != nil {
					c.Error(err)
					c.AbortWithStatusJSON(
						http.StatusInternalServerError,
						map[string]string{"code": "user_confirm didn't passed"},
					)
					return
				}

				res, err := app.Session.InsertUser(
					*app.Auth.UserColl,
					*storage.NewInsertUserData(userUnique.(string), userConfirm.(string)),
				)
				if err != nil {
					c.AbortWithStatus(http.StatusInternalServerError)
				}

				if registerConfig.LoginAfter {
					c.JSON(http.StatusOK, map[string]string{"token": "jwt"})
					return
				}
				c.JSON(http.StatusOK, res)
			})
		}
	}

	return r
}
