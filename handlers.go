package main

import (
	"github.com/gin-gonic/gin"
	"github.com/gofiber/fiber/v2"
	"gouth/authN"
	"gouth/config"
	"gouth/jwt"
	"gouth/storage"
	"net/http"
	"strings"
)

func registerHandler(app *config.App) func(c *gin.Context) {
	return func(c *gin.Context) {
		var regData interface{}

		if err := c.BindJSON(&regData); err != nil {
			c.AbortWithStatusJSON(
				http.StatusBadRequest,
				gin.H{"error": "invalid json"})
			return
		}

		var regConfig = app.Register
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

		rawUserConfirm, err := GetJSONPath(mapKeys["user_confirm"], regData)
		if err != nil {
			c.AbortWithStatusJSON(
				http.StatusBadRequest,
				gin.H{"error": "user_confirm didn't passed"},
			)
			return
		}
		userConfirm := rawUserConfirm.(string)
		if strings.TrimSpace(userConfirm) == "" {
			c.AbortWithStatusJSON(
				http.StatusBadRequest,
				gin.H{"error": "user_confirm can't be blank"},
			)
			return
		}

		// TODO: add a user existence check

		pwHash, err := app.Hash.HashPw(userConfirm)
		if err != nil {
			c.AbortWithStatusJSON(
				http.StatusInternalServerError,
				gin.H{"error": err.Error()})
			return
		}

		usersStorage := app.StorageByFeature["users"]
		res, err := usersStorage.InsertUser(
			*app.UserColl,
			*storage.NewInsertUserData(userUnique, pwHash),
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
	}
}

func authNHandler(app *config.App, authNConfig *config.AuthNConfig) func(c *fiber.Ctx) error {
	passwordBased := func(c *fiber.Ctx) error {
		var authData interface{}

		if err := c.BindJSON(&authData); err != nil {
			c.AbortWithStatusJSON(
				http.StatusBadRequest,
				gin.H{"error": "invalid json"})
			return
		}

		userUnique, err := GetJSONPath(authNConfig.PasswdBased.UserUnique, authData)
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

		rawUserConfirm, err := GetJSONPath(authNConfig.PasswdBased.UserConfirm, authData)
		if err != nil {
			c.AbortWithStatusJSON(
				http.StatusBadRequest,
				gin.H{"error": "user_confirm didn't passed"},
			)
			return
		}
		userConfirm := rawUserConfirm.(string)
		if strings.TrimSpace(userConfirm) == "" {
			c.AbortWithStatusJSON(
				http.StatusBadRequest,
				gin.H{"error": "user_confirm can't be blank"},
			)
			return
		}

		// TODO: add a user existence check
		usersStorage := app.StorageByFeature["users"]
		pw, err := usersStorage.GetUserPassword(*app.UserColl, userUnique)
		if err != nil {
			c.AbortWithStatusJSON(
				http.StatusInternalServerError,
				gin.H{"error": err.Error()})
			return
		}

		isMatch, err := app.Hash.ComparePw(userConfirm, pw.(string))
		if err != nil {
			c.AbortWithStatusJSON(
				http.StatusInternalServerError,
				gin.H{"error": err.Error()})
			return
		}

		if isMatch {
			token := jwt.IssueToken()
			c.JSON(http.StatusOK, gin.H{"token": token})
		} else {
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				gin.H{"error": "invalid data"})
			return
		}
	}

	switch authNConfig.Type {
	case authN.PasswordBased:
		return passwordBased
	default:
		return nil
	}
}
