package main

import (
	"github.com/gin-gonic/gin"
	"gouth/jwt"
	"gouth/storage"
	"net/http"
	"strings"
)

// initRouter initializes router and creates routes for each application
func initRouter() *gin.Engine {
	r := gin.Default()
	v := r.Group("v" + conf.APIVersion)

	for _, app := range conf.Apps {
		appR := v.Group(app.PathPrefix)
		{
			appR.POST("/register", func(c *gin.Context) {
				var regData interface{}

				if err := c.BindJSON(&regData); err != nil {
					c.AbortWithStatusJSON(
						http.StatusBadRequest,
						gin.H{"error": "invalid json"})
					return
				}

				var regConfig = app.Main.Register
				mapKeys := regConfig.Fields

				userUnique, err := GetJSONPath(mapKeys["user_unique"], regData)
				if err != nil {
					c.AbortWithStatusJSON(
						http.StatusBadRequest,
						gin.H{"error": "user_unique didn't passed"},
					)
					return
				}
				if userUnique, ok := userUnique.(string); ok {
					if strings.TrimSpace(userUnique) == "" {
						c.AbortWithStatusJSON(
							http.StatusBadRequest,
							gin.H{"error": "user_unique can't be blank"},
						)
						return
					}
				}

				userConfirm, err := GetJSONPath(mapKeys["user_confirm"], regData)
				if err != nil {
					c.AbortWithStatusJSON(
						http.StatusBadRequest,
						gin.H{"error": "user_confirm didn't passed"},
					)
					return
				}
				if userConfirm, ok := userConfirm.(string); ok {
					if strings.TrimSpace(userConfirm) == "" {
						c.AbortWithStatusJSON(
							http.StatusBadRequest,
							gin.H{"error": "user_confirm can't be blank"},
						)
						return
					}
				}

				// TODO: add a user existence check

				res, err := app.Session.InsertUser(
					*app.Main.UserColl,
					*storage.NewInsertUserData(userUnique, userConfirm),
				)
				if err != nil {
					c.AbortWithStatusJSON(
						http.StatusInternalServerError,
						gin.H{"error": err.Error()})
					return
				}

				if regConfig.LoginAfter {
					token := jwt.IssueToken()

					c.JSON(http.StatusOK, gin.H{"token": token})
					return
				}

				c.JSON(http.StatusOK, gin.H{"id": res})
			})
		}
	}

	return r
}
