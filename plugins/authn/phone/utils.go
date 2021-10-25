package phone

import (
	"aureole/internal/identity"
	crand "crypto/rand"
	"github.com/gofiber/fiber/v2"
	"math/big"
)

func sendError(c *fiber.Ctx, statusCode int, message string) error {
	return c.Status(statusCode).JSON(&fiber.Map{
		"success": false,
		"message": message,
	})
}

func isCredential(trait *identity.Trait) bool {
	return trait.IsCredential && trait.IsRequired && trait.IsUnique
}

func getRandomString(length int, alphabet string) (string, error) {
	ret := make([]byte, length)
	for i := 0; i < length; i++ {
		num, err := crand.Int(crand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			return "", err
		}
		ret[i] = alphabet[num.Int64()]
	}

	return string(ret), nil
}
