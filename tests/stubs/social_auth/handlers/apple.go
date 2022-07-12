package handlers

import (
	"bytes"
	"context"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jwt"
	"io/ioutil"
	"log"
	"net/http"
)

func AppleAuthUrlHandler(c *fiber.Ctx) error {
	state := c.Query("state")
	redirectUri := c.Query("redirect_uri")
	log.Println("Got new request with:", state, redirectUri)

	var buf bytes.Buffer
	buf.WriteString(redirectUri)

	jsonData := []byte(fmt.Sprintf(`{"state": "%s", "code": "12345"}`, state))
	ctx := context.Background()
	r, err := http.NewRequestWithContext(ctx, http.MethodPost, redirectUri, bytes.NewBuffer(jsonData))
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}

	r.Header.Set("Content-Type", "application/json; charset=UTF-8")
	res, err := http.DefaultClient.Do(r)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}

	defer res.Body.Close()
	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}

	c.Response().SetBodyRaw(bodyBytes)
	for _, cookie := range res.Cookies() {
		c.Cookie(&fiber.Cookie{
			Name:  cookie.Name,
			Value: cookie.Value,
		})
	}
	c.Set("access", res.Header.Get("access"))
	return c.SendStatus(res.StatusCode)
}

func AppleTokenHandler(c *fiber.Ctx) error {
	token := jwt.New()

	if err := token.Set(jwt.SubjectKey, "123456"); err != nil {
		return err
	}
	if err := token.Set(jwt.AudienceKey, "123456"); err != nil {
		return err
	}
	if err := token.Set("email", "example@gmail.com"); err != nil {
		return err
	}

	keySet, err := jwk.ReadFile("/resources/keys.json")
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}

	key, ok := keySet.Get(0)
	if !ok {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}
	signed, err := jwt.Sign(token, jwa.RS256, key)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}

	return c.JSON(fiber.Map{
		"id_token": string(signed),
	})
}
