package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jwt"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

func createJWT(payload map[string]interface{}, keySet jwk.Set, exp int) (string, error) {
	token, err := newToken(payload, exp)
	if err != nil {
		return "", err
	}

	signedToken, err := signToken(keySet, token)
	if err != nil {
		return "", err
	}
	return string(signedToken), nil
}

func newToken(payload map[string]interface{}, exp int) (jwt.Token, error) {
	predefinedKeys := []string{
		jwt.JwtIDKey,
		jwt.IssuerKey,
		jwt.ExpirationKey,
		jwt.NotBeforeKey,
		jwt.IssuedAtKey,
		"type",
	}
	for _, k := range predefinedKeys {
		delete(payload, k)
	}

	token := jwt.New()

	jti, err := gonanoid.New(16)
	if err != nil {
		return nil, err
	}
	err = token.Set(jwt.JwtIDKey, jti)
	if err != nil {
		return nil, err
	}
	err = token.Set(jwt.IssuerKey, "Aureole")
	if err != nil {
		return nil, err
	}
	err = token.Set(jwt.ExpirationKey, time.Now().Add(time.Duration(exp)*time.Second).Unix())
	if err != nil {
		return nil, err
	}
	err = token.Set(jwt.NotBeforeKey, time.Now().Unix())
	if err != nil {
		return nil, err
	}
	err = token.Set(jwt.IssuedAtKey, time.Now().Unix())
	if err != nil {
		return nil, err
	}

	for k, v := range payload {
		err := token.Set(k, v)
		if err != nil {
			return nil, err
		}
	}
	return token, err
}

func signToken(keySet jwk.Set, token jwt.Token) ([]byte, error) {
	for it := keySet.Iterate(context.Background()); it.Next(context.Background()); {
		pair := it.Pair()
		key := pair.Value.(jwk.Key)

		if key.KeyUsage() == "sig" {
			var signAlg jwa.SignatureAlgorithm
			if err := signAlg.Accept(key.Algorithm()); err != nil {
				return nil, err
			}
			return jwt.Sign(token, signAlg, key)
		}
	}
	return nil, errors.New("key set don't contain sig key")
}

func getFromJWT(token jwt.Token, name string, value interface{}) error {
	data, ok := token.Get(name)
	if !ok {
		return fmt.Errorf("cannot get '%s' field from token", name)
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return json.Unmarshal(bytes, value)
}
