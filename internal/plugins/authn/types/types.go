package types

import (
	coll "aureole/internal/collections"
	"aureole/internal/identity"
	_interface "aureole/internal/state/interface"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/mitchellh/mapstructure"
)

type Authenticator interface {
	Init(app _interface.AppState) error
}

type Input struct {
	Id         interface{}            `mapstructure:"id"`
	Username   interface{}            `mapstructure:"username"`
	Phone      interface{}            `mapstructure:"phone"`
	Email      interface{}            `mapstructure:"email"`
	Password   interface{}            `mapstructure:"password"`
	Additional map[string]interface{} `mapstructure:",remain"`
}

func NewInput(c *fiber.Ctx) (*Input, error) {
	var (
		rawInput interface{}
		input    *Input
	)
	if err := c.BodyParser(&rawInput); err != nil {
		return nil, err
	}
	if err := mapstructure.Decode(rawInput, &input); err != nil {
		return nil, err
	}
	return input, nil
}

func (c *Input) Init(identity *identity.Identity, collMap map[string]coll.FieldSpec, validate bool) error {
	c.setDefaults(collMap)

	if validate {
		if err := c.validate(identity, collMap); err != nil {
			return err
		}
	}

	return nil
}

func (c *Input) setDefaults(collMap map[string]coll.FieldSpec) {
	if c.Username == nil {
		c.Username = collMap["username"].Default
	}

	if c.Phone == nil {
		c.Phone = collMap["phone"].Default
	}

	if c.Email == nil {
		c.Email = collMap["email"].Default
	}

	for name := range c.Additional {
		if c.Additional[name] == nil {
			c.Additional[name] = collMap[name].Default
		}
	}
}

func (c *Input) validate(identity *identity.Identity, collMap map[string]coll.FieldSpec) error {
	if c.Username == nil && identity.Username.IsRequired {
		return errors.New("username is required, but isn't passed")
	}

	if c.Phone == nil && identity.Phone.IsRequired {
		return errors.New("phone is required, but isn't passed")
	}

	if c.Email == nil && identity.Email.IsRequired {
		return errors.New("email is required, but isn't passed")
	}

	for name, val := range c.Additional {
		if identity.Additional[name].IsInternal && val != collMap[name].Default {
			c.Additional[name] = nil
		}

		if val == nil && identity.Additional[name].IsRequired {
			return fmt.Errorf("%s is required, but not passed", name)
		}
	}

	return nil
}
