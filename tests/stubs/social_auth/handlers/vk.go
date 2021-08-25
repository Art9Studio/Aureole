package handlers

import (
	"bytes"
	"github.com/gofiber/fiber/v2"
	"log"
	"net/url"
)

func VkAuthUrlHandler(c *fiber.Ctx) error {
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

func VkTokenHandler(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"access_token":  "abcd",
		"token_type":    "abcd",
		"refresh_token": "abcd",
		"expires_in":    200,
		"email":         "example@gmail.com",
		"user_id":       "123456",
	})
}

func VkUserDataHandler(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"response": []map[string]interface{}{
		{"first_name": "John"},
	}})
}
