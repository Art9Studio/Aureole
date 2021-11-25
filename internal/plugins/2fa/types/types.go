package types

import (
	"aureole/internal/identity"
	"aureole/internal/plugins"
	"github.com/gofiber/fiber/v2"
)

type (
	MFAFunc func(fiber.Ctx) (*identity.Credential, fiber.Map, error)

	SecondFactor interface {
		plugins.MetaDataGetter
		IsEnabled(cred *identity.Credential, provider string) (bool, error)
		Init2FA(cred *identity.Credential, provider string, c fiber.Ctx) (fiber.Map, error)
		Verify() MFAFunc
	}
)
