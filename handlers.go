package main

//func registerHandler(app *configs.App) func(c *gin.Context) {
//	return func(c *gin.Context) {
//		var regData interface{}
//
//		if err := c.BindJSON(&regData); err != nil {
//			c.AbortWithStatusJSON(
//				http.StatusBadRequest,
//				gin.H{"error": "invalid json"})
//			return
//		}
//
//		var regConfig = app.Register
//		mapKeys := regConfig.Fields
//
//		userUnique, err := GetJSONPath(mapKeys["identity"], regData)
//		if err != nil {
//			c.AbortWithStatusJSON(
//				http.StatusBadRequest,
//				gin.H{"error": "identity didn't passed"},
//			)
//			return
//		}
//		if userUnique, ok := userUnique.(string); ok {
//			if strings.TrimSpace(userUnique) == "" {
//				c.AbortWithStatusJSON(
//					http.StatusBadRequest,
//					gin.H{"error": "identity can't be blank"},
//				)
//				return
//			}
//		}
//
//		rawUserConfirm, err := GetJSONPath(mapKeys["password"], regData)
//		if err != nil {
//			c.AbortWithStatusJSON(
//				http.StatusBadRequest,
//				gin.H{"error": "password didn't passed"},
//			)
//			return
//		}
//		userConfirm := rawUserConfirm.(string)
//		if strings.TrimSpace(userConfirm) == "" {
//			c.AbortWithStatusJSON(
//				http.StatusBadRequest,
//				gin.H{"error": "password can't be blank"},
//			)
//			return
//		}
//
//		// TODO: add a user existence check
//
//		pwHash, err := app.Hash.HashPw(userConfirm)
//		if err != nil {
//			c.AbortWithStatusJSON(
//				http.StatusInternalServerError,
//				gin.H{"error": err.Error()})
//			return
//		}
//
//		usersStorage := app.StorageByFeature["identity"]
//		res, err := usersStorage.InsertIdentity(
//			*app.UserColl,
//			*storage.NewInsertUserData(userUnique, pwHash),
//		)
//		if err != nil {
//			c.AbortWithStatusJSON(
//				http.StatusInternalServerError,
//				gin.H{"error": err.Error()})
//			return
//		}
//
//		if regConfig.LoginAfter {
//			token := jwt.IssueToken()
//
//			c.JSON(http.StatusOK, gin.H{"token": token})
//			return
//		}
//
//		c.JSON(http.StatusOK, gin.H{"id": res})
//	}
//}
//
//func authNHandler(app *configs.App, authnConfig *configs.AuthnConfig) func(c *fiber.Ctx) error {
//
//	switch authnConfig.Kind {
//	case types.PasswordBased:
//		return passwordBased
//	default:
//		return nil
//	}
//}
