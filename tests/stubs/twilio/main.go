package main

import (
	"github.com/gofiber/fiber/v2"
	"log"
)

type Message struct {
	Body string
	From string
	To   string
}

var messages []*Message

func main() {
	app := fiber.New()

	app.Get("/ping", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})
	app.Post("/2010-04-01/Accounts/123456/Messages.json", func(c *fiber.Ctx) error {
		m := new(Message)
		if err := c.BodyParser(m); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":  400,
				"message": "can't parse message",
			})
		}
		messages = append(messages, m)
		return c.SendStatus(fiber.StatusOK)
	})
	app.Get("/2010-04-01/Accounts/123456/Messages.json", func(c *fiber.Ctx) error {
		return c.JSON(messages)
	})
	app.Delete("/2010-04-01/Accounts/123456/Messages.json", func(c *fiber.Ctx) error {
		messages = nil
		return c.SendStatus(fiber.StatusOK)
	})

	if err := app.ListenTLS(":443", "certs/server.crt", "certs/server.key"); err != nil {
		log.Panicf("starting server: %v", err)
	}
}
