package handlers

import (
	"bytes"
	"github.com/gofiber/fiber/v2"
	"log"
	"net/url"
)

func FacebookAuthUrlHandler(c *fiber.Ctx) error {
	state := c.Query("state")
	redirectUri := c.Query("redirect_uri")
	log.Println("Got new request with:", state, redirectUri)

	var buf bytes.Buffer
	buf.WriteString(redirectUri)

	v := url.Values{
		"state": {state},
		"code":  {"123456"},
	}

	buf.WriteByte('?')
	buf.WriteString(v.Encode())

	var redirect = buf.String()

	log.Println("Redirecting to", redirect)
	return c.Redirect(redirect)
}

func FacebookTokenHandler(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"access_token":  "abcd",
		"token_type":    "abcd",
		"refresh_token": "abcd",
		"expires_in":    200,
	})
}

func FacebookUserDataHandler(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"id":    "123456",
		"email": "example@gmail.com",
	})
}
