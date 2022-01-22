package plugins

import (
	"aureole/internal/configs"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

var SecondFactorRepo = createRepository()

type (
	SecondFactorAdapter interface {
		Create(*configs.SecondFactor) SecondFactor
	}

	SecondFactor interface {
		MetaDataGetter
		IsEnabled(cred *Credential, provider string) (bool, error)
		Init2FA(cred *Credential, provider string, c fiber.Ctx) (fiber.Map, error)
		Verify() MFAVerifyFunc
	}

	MFAVerifyFunc func(fiber.Ctx) (*Credential, fiber.Map, error)
)

func NewSecondFactor(conf *configs.SecondFactor) (SecondFactor, error) {
	a, err := SecondFactorRepo.Get(conf.Type)
	if err != nil {
		return nil, err
	}

	adapter, ok := a.(SecondFactorAdapter)
	if !ok {
		return nil, fmt.Errorf("trying to cast adapter was failed: %v", err)
	}

	return adapter.Create(conf), nil
}
