package identity

import (
	"errors"
	"fmt"
	"github.com/mitchellh/mapstructure"
)

type (
	ManagerI interface {
		OnUserAuthenticated(cred *Credential, identity *Identity, provider string) (*Identity, error)
		Register(cred *Credential, identity *Identity, provider string) (*Identity, error)
		On2FA(cred *Credential, data map[string]interface{}) error
		GetData(cred *Credential, provider string, name string) (interface{}, error) // описать поля, которые можем получить
		Update(cred *Credential, identity *Identity, provider string) (*Identity, error)
		CheckFeaturesAvailable(features []string) error
	}

	Manager struct {
	}

	Credential struct {
		Name  string
		Value string
	}

	Identity struct {
		ID            interface{}
		Email         string
		Phone         string
		Username      string
		EmailVerified bool
		PhoneVerified bool
		Additional    map[string]interface{}
	}
)

const (
	ID             = "id"
	SocialID       = "social_id"
	Email          = "email"
	Phone          = "phone"
	Username       = "username"
	EmailVerified  = "email_verified"
	PhoneVerified  = "phone_verified"
	Password       = "password"
	SecondFactorID = "2fa_id"
	AuthnProvider  = "provider"
	UserData       = "user_data"
)

var features = map[string]bool{
	"on_register": true,
	"2fa":         true,
	"get_data":    true,
	"update":      true,
}

func NewIdentity(data map[string]interface{}) (*Identity, error) {
	i := &Identity{}
	if err := mapstructure.Decode(data, i); err != nil {
		return nil, err
	}
	return i, nil
}

func (i *Identity) AsMap() map[string]interface{} {
	identityMap := map[string]interface{}{
		ID:            i.ID,
		Email:         i.Email,
		Phone:         i.Phone,
		Username:      i.Username,
		EmailVerified: i.EmailVerified,
		PhoneVerified: i.PhoneVerified,
	}

	for k, v := range i.Additional {
		identityMap[k] = v
	}

	return identityMap
}

func Create() (*Manager, error) {
	return &Manager{}, nil
}

func (*Manager) OnUserAuthenticated(cred *Credential, i *Identity, _ string) (*Identity, error) {

	// check if user exists by cred
	var exist = true
	if exist {
		// get all data from db by cred
		return &Identity{}, nil
	} else if (cred.Name == "email" && i.EmailVerified) || (cred.Name == "phone" && i.PhoneVerified) {

		fmt.Printf("OnUserAuthenticated\n Identity -  %#v\n", i)

		// insert new user in db

		return i, nil
	}

	return nil, errors.New("user doesn't exists")
}

func (*Manager) Register(_ *Credential, i *Identity, _ string) (*Identity, error) {

	// check if user exists by cred
	var exist bool
	if exist {
		return nil, errors.New("user already exists")
	}

	fmt.Printf("Register\n Identity -  %#v\n", i)

	// save all data to db and return entity

	return i, nil
}

func (*Manager) On2FA(*Credential, map[string]interface{}) error {
	return nil
}

func (*Manager) GetData(_ *Credential, _, output string) (interface{}, error) {

	// check if user exists by cred
	// get some data from db by cred

	switch output {
	case "password":
		return "dummy_password", nil
	case "username":
		return "dummy_username", nil
	}

	return nil, nil
}

func (*Manager) Update(_ *Credential, i *Identity, _ string) (*Identity, error) {

	// get entity by cred
	// update entity and return it

	return i, errors.New("can't determine type of updated data")
}

func (*Manager) CheckFeaturesAvailable(requiredFeatures []string) error {
	for _, f := range requiredFeatures {
		if available, ok := features[f]; !ok || !available {
			return fmt.Errorf("feature %s hasn't implemented", f)
		}
	}
	return nil
}
