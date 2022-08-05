package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strconv"

	"github.com/fatih/structs"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/mitchellh/mapstructure"
)

type (
	IDManager struct {
		pool       *pgxpool.Pool
		privateSet jwk.Set
		publicSet  jwk.Set
		jwtExp     int
		features   map[string]bool
	}

	Credential struct {
		Name  string `mapstructure:"name" json:"name"`
		Value string `mapstructure:"value" json:"value"`
	}

	Identity struct {
		ID            interface{}            `mapstructure:"id,omitempty" json:"id,omitempty"`
		Email         string                 `mapstructure:"email,omitempty" json:"email,omitempty"`
		Phone         string                 `mapstructure:"phone,omitempty" json:"phone,omitempty"`
		Username      string                 `mapstructure:"username,omitempty" json:"username,omitempty"`
		EmailVerified bool                   `mapstructure:"email_verified" json:"email_verified"`
		PhoneVerified bool                   `mapstructure:"phone_verified" json:"phone_verified"`
		Additional    map[string]interface{} `mapstructure:"additional,omitempty" json:"additional,omitempty"`
	}
)

func newIDManager() (*IDManager, error) {
	connStr, ok := os.LookupEnv("DATABASE_URL")
	if !ok {
		return nil, errors.New("cannot find database connection url")
	}
	pool, err := pgxpool.Connect(context.Background(), connStr)
	if err != nil {
		return nil, err
	}

	filepath, ok := os.LookupEnv("CRYPTO_KEYS_PATH")
	if !ok {
		return nil, errors.New("cannot find crypto keys path")
	}
	privSet, pubSet, err := initKeySets(filepath)
	if err != nil {
		return nil, err
	}

	jwtExp, ok := os.LookupEnv("JWT_EXPIRE")
	if !ok {
		return nil, errors.New("cannot find jwt expire time")
	}
	exp, err := strconv.Atoi(jwtExp)
	if err != nil {
		return nil, err
	}

	manager := &IDManager{
		pool:       pool,
		privateSet: privSet,
		publicSet:  pubSet,
		jwtExp:     exp,
		features: map[string]bool{
			"Register":            true,
			"OnUserAuthenticated": true,
			"OnMFA":               true,
			"GetData":             true,
			"GetMFAData":          true,
			"Update":              true,
		},
	}
	return manager, nil
}

func initKeySets(filepath string) (privSet jwk.Set, pubSet jwk.Set, err error) {
	if _, err = os.Stat(filepath); err != nil {
		return nil, nil, err
	}
	keyBytes, err := os.ReadFile(filepath)
	if err != nil {
		return nil, nil, err
	}

	privSet, err = jwk.Parse(keyBytes)
	if err != nil {
		return nil, nil, err
	}
	pubSet, err = jwk.PublicSetOf(privSet)
	if err != nil {
		return nil, nil, err
	}

	return privSet, pubSet, nil
}

func newIdentity(data map[string]interface{}) (*Identity, error) {
	ident := &Identity{}
	if err := mapstructure.Decode(data, ident); err != nil {
		return nil, err
	}

	oauth2Str, ok := ident.Additional["social_providers"]
	if ok {
		oauth2Bytes, err := json.Marshal(oauth2Str)
		if err != nil {
			return nil, err
		}

		var oauth2Data map[string]interface{}
		err = json.Unmarshal(oauth2Bytes, &oauth2Data)
		if err != nil {
			return nil, err
		}

		if len(oauth2Data) != 0 {
			ident.Additional["social_providers"] = oauth2Data
		} else {
			delete(oauth2Data, "social_providers")
		}
	}

	return ident, nil
}

func (i *Identity) AsMap() map[string]interface{} {
	structsConf := structs.New(i)
	structsConf.TagName = "mapstructure"
	return structsConf.Map()
}
