package pem

import (
	"aureole/internal/core"
	"context"
	"crypto/x509"
	pemLib "encoding/pem"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/lestrrat-go/jwx/jwk"
)

func getJwkKeys(p *pem) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		return c.Status(http.StatusOK).JSON(p.publicSet)
	}
}

func getPemKeys(p *pem) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		pemKeySet := map[string]string{}

		for it := p.publicSet.Iterate(context.Background()); it.Next(context.Background()); {
			pair := it.Pair()
			key := pair.Value.(jwk.Key)

			var rawKey interface{}
			if err := key.Raw(&rawKey); err != nil {
				return core.SendError(c, http.StatusInternalServerError, err.Error())
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
					return core.SendError(c, http.StatusInternalServerError, err.Error())
				}
			}

			pemKey := pemLib.EncodeToMemory(&pemLib.Block{
				Type:  "PUBLIC KEY",
				Bytes: pemData,
			})
			if pemKey == nil {
				return core.SendError(c, http.StatusInternalServerError, "cannot get pem from jwk")
			}

			pemKeySet[key.KeyID()] = string(pemKey)
		}

		return c.Status(http.StatusOK).JSON(pemKeySet)
	}
}
