package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"log"
	"social-auth/handlers"
)

func main() {
	app := fiber.New()

	app.Use(logger.New())

	app.Get("/ping", func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusOK)
	})

	app.Get("/o/oauth2/auth", handlers.GoogleAuthUrlHandler)
	app.Post("/token", handlers.GoogleTokenHandler)

	app.Get("/auth/authorize", handlers.AppleAuthUrlHandler)
	app.Post("/auth/token", handlers.AppleTokenHandler)

	app.Get("/v3.2/dialog/oauth", handlers.FacebookAuthUrlHandler)
	app.Post("/v3.2/oauth/access_token", handlers.FacebookTokenHandler)
	app.Get("/me", handlers.FacebookUserDataHandler)

	app.Get("/authorize", handlers.VkAuthUrlHandler)
	app.Post("/access_token", handlers.VkTokenHandler)
	app.Get("/method/users.get", handlers.VkUserDataHandler)

	if err := app.ListenTLS(":443", "certs/server.crt", "certs/server.key"); err != nil {
		log.Panicf("starting server: %v", err)
	}
}
