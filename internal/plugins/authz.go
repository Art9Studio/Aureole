package plugins

import (
	"aureole/internal/configs"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/mitchellh/mapstructure"
)

var AuthZRepo = createRepository()

// AuthZAdapter defines methods for authorization plugins
type (
	AuthZAdapter interface {
		// Create returns desired authorization depends on the given config
		Create(*configs.Authz) Authorizer
	}

	Authorizer interface {
		MetaDataGetter
		GetNativeQueries() map[string]string
		Authorize(*fiber.Ctx, *Payload) error
	}

	Payload struct {
		ID         interface{}            `mapstructure:"id"`
		SocialID   interface{}            `mapstructure:"social_id"`
		Username   string                 `mapstructure:"username"`
		Phone      string                 `mapstructure:"phone"`
		Email      string                 `mapstructure:"email"`
		UserData   interface{}            `mapstructure:"user_data"`
		Additional map[string]interface{} `mapstructure:",remain"`
		// NativeQ    func(queryName string, args ...interface{}) string
	}
)

// NewAuthZ returns desired authorizer depends on the given config
func NewAuthZ(conf *configs.Authz) (Authorizer, error) {
	a, err := AuthZRepo.Get(conf.Type)
	if err != nil {
		return nil, err
	}

	adapter, ok := a.(AuthZAdapter)
	if !ok {
		return nil, fmt.Errorf("trying to cast adapter was failed: %v", err)
	}

	return adapter.Create(conf), nil
}

func NewPayload(data map[string]interface{}) (*Payload, error) {
	p := &Payload{}
	if err := mapstructure.Decode(data, p); err != nil {
		return nil, err
	}
	return p, nil
}
