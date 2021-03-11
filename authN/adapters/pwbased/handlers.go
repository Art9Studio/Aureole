package pwbased

import "github.com/gofiber/fiber/v2"

func Auth() func  {
	(c *fiber.Ctx) error {
		var authInput interface{}

		if err := c.BodyParser(&authInput); err != nil {
		return c.Status(400).JSON(&fiber.Map{
		"success": false,
		"message": err,
	})
	}

		userUnique, err := jsonpath.GetJSONPath(authNConfig.PasswdBased.UserUnique, authInput)
		if err != nil {
		c.AbortWithStatusJSON(
		http.StatusBadRequest,
		gin.H{"error": "identity didn't passed"},
	)
		return
	}
		if userUnique, ok := userUnique.(string); ok {
		if strings.TrimSpace(userUnique) == "" {
		c.AbortWithStatusJSON(
		http.StatusBadRequest,
		gin.H{"error": "identity can't be blank"},
	)
		return
	}
	}

		rawUserConfirm, err := jsonpath.GetJSONPath(authNConfig.PasswdBased.UserConfirm, authInput)
		if err != nil {
		c.AbortWithStatusJSON(
		http.StatusBadRequest,
		gin.H{"error": "password didn't passed"},
	)
		return
	}
		userConfirm := rawUserConfirm.(string)
		if strings.TrimSpace(userConfirm) == "" {
		c.AbortWithStatusJSON(
		http.StatusBadRequest,
		gin.H{"error": "password can't be blank"},
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
}
