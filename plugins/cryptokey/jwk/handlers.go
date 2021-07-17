package jwk

import (
	"context"
	"crypto/x509"
	"encoding/pem"

	"github.com/gofiber/fiber/v2"
	"github.com/lestrrat-go/jwx/jwk"
)

func GetJwkKeys(j *Jwk) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(j.publicSet)
	}
}

func GetPemKeys(j *Jwk) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		pemKeySet := map[string]string{}

		for it := j.publicSet.Iterate(context.Background()); it.Next(context.Background()); {
			pair := it.Pair()
			key := pair.Value.(jwk.Key)

			var rawKey interface{}
			if err := key.Raw(&rawKey); err != nil {
				return sendError(c, fiber.StatusInternalServerError, err.Error())
			}

			var (
				pemData []byte
				err     error
			)
			if keyBytes, ok := rawKey.([]byte); ok {
				pemData = keyBytes
			} else {
				pemData, err = x509.MarshalPKIXPublicKey(rawKey)
				if err != nil {
					return sendError(c, fiber.StatusInternalServerError, err.Error())
				}
			}

			pemKey := pem.EncodeToMemory(&pem.Block{
				Type:  "PUBLIC KEY",
				Bytes: pemData,
			})
			if pemKey == nil {
				return sendError(c, fiber.StatusInternalServerError, "cannot get pem from jwk")
			}

			pemKeySet[key.KeyID()] = string(pemKey)
		}

		return c.Status(fiber.StatusOK).JSON(pemKeySet)
	}
}

func sendError(c *fiber.Ctx, statusCode int, message string) error {
	return c.Status(statusCode).JSON(&fiber.Map{
		"success": false,
		"message": message,
	})
}
