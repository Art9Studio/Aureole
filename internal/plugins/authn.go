package plugins

import (
	"aureole/internal/configs"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

var AuthNRepo = createRepository()

type (
	// AuthNAdapter defines methods for authentication plugins
	AuthNAdapter interface {
		// Create returns desired authentication Authenticator depends on the given config
		Create(*configs.Authn) Authenticator
	}

	Authenticator interface {
		MetaDataGetter
		Login() AuthNLoginFunc
	}

	AuthNResult struct {
		Cred       *Credential
		Identity   *Identity
		Provider   string
		Additional map[string]interface{}
	}

	AuthNLoginFunc func(fiber.Ctx) (*AuthNResult, error)
)

// NewAuthN returns desired Authenticator depends on the given config
func NewAuthN(conf *configs.Authn) (Authenticator, error) {
	a, err := AuthNRepo.Get(conf.Type)
	if err != nil {
		return nil, err
	}

	adapter, ok := a.(AuthNAdapter)
	if !ok {
		return nil, fmt.Errorf("trying to cast adapter was failed: %v", err)
	}

	return adapter.Create(conf), nil
}
