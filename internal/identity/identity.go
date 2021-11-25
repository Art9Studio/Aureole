package identity

import (
	"errors"
	"fmt"
	"github.com/mitchellh/mapstructure"
)

type (
	ManagerI interface {
		OnUserAuthenticated(cred *Credential, identity *Identity, provider string, additional map[string]interface{}) (map[string]interface{}, error)
		OnRegister(cred *Credential, identity *Identity, provider string, additional map[string]interface{}) (map[string]interface{}, error)
		On2FA(cred *Credential, data map[string]interface{}) error
		GetData(cred *Credential, provider string, name string) (interface{}, error) // описать поля, которые можем получить
		Update(cred *Credential, provider string, data map[string]interface{}) (map[string]interface{}, error)
		CheckFeaturesAvailable(features []string) error
	}

	Manager struct {
	}

	Credential struct {
		Name  string
		Value string
	}

	Identity struct {
		Id            interface{}
		Email         string
		Phone         string
		Username      string
		EmailVerified bool
		PhoneVerified bool
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

func Create() (*Manager, error) {
	return &Manager{}, nil
}

func (*Manager) OnUserAuthenticated(cred *Credential, i *Identity, _ string, additional map[string]interface{}) (map[string]interface{}, error) {

	// check if user exists by cred
	var exist = true
	if exist {
		// get all data from db by cred
		return map[string]interface{}{}, nil
	} else if (cred.Name == "email" && i.EmailVerified) || (cred.Name == "phone" && i.PhoneVerified) {

		fmt.Printf("OnUserAuthenticated\n Identity -  %#v\n", i)
		fmt.Printf("Additional %#v\n", additional)

		// insert new user in db

		return map[string]interface{}{
			"id":         i.Id,
			"username":   i.Username,
			"email":      i.Email,
			"phone":      i.Phone,
			"additional": additional,
		}, nil
	}

	return nil, errors.New("user doesn't exists")
}

func (*Manager) OnRegister(_ *Credential, i *Identity, _ string, additional map[string]interface{}) (map[string]interface{}, error) {

	// check if user exists by cred
	var exist bool
	if exist {
		return nil, errors.New("user already exists")
	}

	fmt.Printf("OnRegister\n Identity -  %#v\n", i)
	fmt.Printf("Additional %#v\n", additional)

	// save all data to db and return entity

	return map[string]interface{}{
		"id":         i.Id,
		"username":   i.Username,
		"email":      i.Email,
		"phone":      i.Phone,
		"additional": additional,
	}, nil
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

func (*Manager) Update(_ *Credential, _ string, fields map[string]interface{}) (map[string]interface{}, error) {

	// get entity by cred
	// update entity and return it

	return fields, errors.New("can't determine type of updated data")
}

func (*Manager) CheckFeaturesAvailable(requiredFeatures []string) error {
	for _, f := range requiredFeatures {
		if available, ok := features[f]; !ok || !available {
			return fmt.Errorf("feature %s hasn't implemented", f)
		}
	}
	return nil
}
