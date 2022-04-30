package plugins

import (
	"github.com/gofiber/fiber/v2"
)

type (
	MFA interface {
		MetaDataGetter
		OpenAPISpecGetter
		IsEnabled(cred *Credential) (bool, error)
		Init2FA() MFAInitFunc
		Verify() MFAVerifyFunc
	}

	MFAVerifyFunc func(fiber.Ctx) (cred *Credential, errorData fiber.Map, err error)
	MFAInitFunc   func(fiber.Ctx) (mfaData fiber.Map, err error)
)
