package handlers

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"github.com/gofiber/fiber/v2"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwt"
	"log"
	"net/url"
)

func GoogleAuthUrlHandler(c *fiber.Ctx) error {
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

func GoogleTokenHandler(c *fiber.Ctx) error {
	token := jwt.New()

	if err := token.Set(jwt.SubjectKey, "123456"); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}
	if err := token.Set("email", "example@gmail.com"); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}

	signed, err := jwt.Sign(token, jwa.RS256, key)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}

	return c.JSON(fiber.Map{
		"access_token":  "abcd",
		"token_type":    "abcd",
		"refresh_token": "abcd",
		"expires_in":    200,
		"id_token":      string(signed),
	})
}
