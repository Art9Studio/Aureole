package pwbased

import (
	"aureole/internal/identity"
	storageT "aureole/internal/plugins/storage/types"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"net/url"
)

func sendError(c *fiber.Ctx, statusCode int, message string) error {
	return c.Status(statusCode).JSON(&fiber.Map{
		"success": false,
		"message": message,
	})
}

func getCredField(i *identity.Identity, iData *storageT.IdentityData) (string, interface{}, error) {
	var credName string
	credVals := map[string]interface{}{}

	if iData.Username != nil && isCredential(i.Username) {
		credVals["username"] = iData.Username
		credName = "username"
	}

	if iData.Email != nil && isCredential(i.Email) {
		credVals["email"] = iData.Email
		credName = "email"
	}

	if iData.Phone != nil && isCredential(i.Phone) {
		credVals["phone"] = iData.Phone
		credName = "phone"
	}

	if l := len(credVals); l != 1 {
		return "", nil, fmt.Errorf("expects 1 credential, %d got", l)
	}

	return credName, credVals[credName], nil
}

func isCredential(trait identity.Trait) bool {
	return trait.IsCredential && trait.IsUnique
}

func initConfirmLink(u *url.URL, token string) string {
	q := u.Query()
	q.Set("token", token)
	u.RawQuery = q.Encode()

	return u.String()
}
