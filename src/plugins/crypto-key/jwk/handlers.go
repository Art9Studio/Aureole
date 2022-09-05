package jwk

import (
	"aureole/internal/core"
	"context"
	"crypto/x509"
	"encoding/pem"
	"net/http"

	"github.com/gofiber/fiber/v2"
	jwx "github.com/lestrrat-go/jwx/jwk"
)

func getJwkKeys(j *jwk) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		return c.JSON(j.publicSet)
	}
}

func getPemKeys(j *jwk) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		pemKeySet := map[string]string{}

		for it := j.publicSet.Iterate(context.Background()); it.Next(context.Background()); {
			pair := it.Pair()
			key := pair.Value.(jwx.Key)

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

			pemKey := pem.EncodeToMemory(&pem.Block{
				Type:  "PUBLIC KEY",
				Bytes: pemData,
			})
			if pemKey == nil {
				return core.SendError(c, http.StatusInternalServerError, "cannot get pem from jwk")
			}

			pemKeySet[key.KeyID()] = string(pemKey)
		}

		return c.JSON(pemKeySet)
	}
}
