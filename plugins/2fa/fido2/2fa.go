package fido2

import (
	"aureole/internal/configs"
	"aureole/internal/core"
	"aureole/internal/plugins"
	"bytes"
	_ "embed"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/duo-labs/webauthn/protocol"
	"github.com/duo-labs/webauthn/webauthn"
	"github.com/go-openapi/spec"
	"github.com/gofiber/fiber/v2"
	"net/http"

	"github.com/mitchellh/mapstructure"
)

const pluginID = "2734"

type (
	mfa struct {
		pluginAPI core.PluginAPI
		rawConf   *configs.SecondFactor
		conf      *config
		idManager plugins.IDManager
		webauthn  *webauthn.WebAuthn
	}

	credentialInput struct {
		Id       interface{} `json:"id,omitempty"`
		Email    string      `json:"email,omitempty"`
		Phone    string      `json:"phone,omitempty"`
		Username string      `json:"username,omitempty"`
	}

	token struct {
		Token string `json:"token"`
	}
)

//go:embed docs/swagger.json
var swaggerJson []byte

func (m *mfa) Init(api core.PluginAPI) (err error) {
	m.pluginAPI = api
	m.conf, err = initConfig(&m.rawConf.Config)
	if err != nil {
		return err
	}

	m.idManager, err = m.pluginAPI.GetIDManager()
	if err != nil {
		return err
	}

	m.webauthn, err = webauthn.New(&webauthn.Config{
		RPDisplayName: "Aureole",
		RPID:          "localhost",
		RPOrigin:      "http://localhost:5500",
	})
	if err != nil {
		return err
	}

	createRoutes(m)
	return nil
}

func (m *mfa) GetMetaData() plugins.Meta {
	return plugins.Meta{
		Type: adapterName,
		Name: m.rawConf.Name,
		ID:   pluginID,
	}
}

func (m *mfa) GetHandlersSpec() (*spec.Paths, spec.Definitions) {
	specs := struct {
		Paths       *spec.Paths
		Definitions spec.Definitions
	}{}
	err := json.Unmarshal(swaggerJson, &specs)
	if err != nil {
		return nil, nil
	}
	return specs.Paths, specs.Definitions
}

func (m *mfa) IsEnabled(cred *plugins.Credential) (bool, error) {
	return m.pluginAPI.Is2FAEnabled(cred, pluginID)
}

func initConfig(rawConf *configs.RawConfig) (*config, error) {
	adapterConf := &config{}
	if err := mapstructure.Decode(rawConf, adapterConf); err != nil {
		return nil, err
	}
	adapterConf.setDefaults()
	return adapterConf, nil
}

func (m *mfa) Init2FA() plugins.MFAInitFunc {
	return func(c fiber.Ctx) (fiber.Map, error) {
		var strToken token
		if err := c.BodyParser(&strToken); err != nil {
			return nil, err
		}
		if strToken.Token == "" {
			return nil, errors.New("token are required")
		}

		token, err := m.pluginAPI.ParseJWT(strToken.Token)
		if err != nil {
			return nil, err
		}
		_, ok := token.Get("provider")
		if !ok {
			return nil, errors.New("cannot get provider from token")
		}
		var cred plugins.Credential
		err = m.pluginAPI.GetFromJWT(token, "credential", &cred)
		if err != nil {
			return nil, errors.New("cannot get credential from jwt: " + err.Error())
		}

		user, err := getUser(&cred, m)
		if err != nil {
			return nil, err
		}

		assertion, sessionData, err := m.webauthn.BeginLogin(user, webauthn.WithUserVerification("discouraged"))
		if err != nil {
			return nil, errors.New("error creating assertion: " + err.Error())
		}

		err = m.pluginAPI.SaveToService(cred.Value+"$fido2$assertion", sessionData, m.pluginAPI.GetAuthSessionExp())
		if err != nil {
			return nil, err
		}
		return fiber.Map{"token": strToken.Token, "assertion": assertion}, nil
	}
}

func (m *mfa) Verify() plugins.MFAVerifyFunc {
	return func(c fiber.Ctx) (*plugins.Credential, fiber.Map, error) {
		var strToken token
		if err := c.BodyParser(&strToken); err != nil {
			return nil, nil, err
		}
		if strToken.Token == "" {
			return nil, nil, errors.New("token are required")
		}

		token, err := m.pluginAPI.ParseJWT(strToken.Token)
		if err != nil {
			return nil, nil, err
		}
		_, ok := token.Get("provider")
		if !ok {
			return nil, nil, errors.New("cannot get provider from token")
		}
		var userCred plugins.Credential
		err = m.pluginAPI.GetFromJWT(token, "credential", &userCred)
		if err != nil {
			return nil, nil, errors.New("cannot get credential from jwt: " + err.Error())
		}

		user, err := getUser(&userCred, m)
		if err != nil {
			return nil, nil, err
		}

		var sessionData webauthn.SessionData
		ok, err = m.pluginAPI.GetFromService(userCred.Value+"$fido2$assertion", &sessionData)
		if err != nil {
			return nil, nil, err
		}
		if !ok {
			return nil, nil, errors.New("session data has expired")
		}

		parsedResponse, err := protocol.ParseCredentialRequestResponseBody(bytes.NewReader(c.Body()))
		cred, err := m.webauthn.ValidateLogin(user, sessionData, parsedResponse)
		if err != nil {
			return nil, nil, err
		}
		if cred.Authenticator.CloneWarning {
			return nil, nil, fmt.Errorf("credential appears to be cloned: %s", err)
		}

		mfaData, err := m.idManager.Get2FAData(&userCred, pluginID)
		if err != nil {
			return nil, nil, err
		}
		rawCred := mfaData.Payload["default"].(map[string]interface{})
		fidoAuthn := rawCred["Authenticator"].(map[string]interface{})
		fidoAuthn["SignCount"] = cred.Authenticator.SignCount

		err = m.idManager.On2FA(&userCred, mfaData)
		if err != nil {
			return nil, nil, err
		}
		return &userCred, nil, nil
	}
}

func createRoutes(m *mfa) {
	routes := []*core.Route{
		{
			Method:  http.MethodPost,
			Path:    startRegistrationURL,
			Handler: startRegistration(m),
		},
		{
			Method:  http.MethodPost,
			Path:    finishRegistrationURL,
			Handler: finishRegistration(m),
		},
	}
	m.pluginAPI.AddAppRoutes(routes)
}

func getCredential(c *credentialInput) (*plugins.Credential, error) {
	if c.Username != "" {
		return &plugins.Credential{
			Name:  "username",
			Value: c.Username,
		}, nil
	}

	if c.Email != "" {
		return &plugins.Credential{
			Name:  "email",
			Value: c.Email,
		}, nil
	}

	if c.Phone != "" {
		return &plugins.Credential{
			Name:  "phone",
			Value: c.Phone,
		}, nil
	}

	return nil, errors.New("credential not found")
}

type user struct {
	identity *plugins.Identity
	mfaData  *plugins.MFAData
}

// WebAuthnID returns the user ID as a byte slice
func (u user) WebAuthnID() []byte {
	buf := make([]byte, binary.MaxVarintLen32)
	binary.PutUvarint(buf, uint64(u.identity.ID.(int32)))
	return buf
}

// WebAuthnName returns the user's username
func (u user) WebAuthnName() string {
	return *u.identity.Username
}

// WebAuthnDisplayName returns the user's display name
func (u user) WebAuthnDisplayName() string {
	return *u.identity.Username
}

// WebAuthnIcon returns the user's icon
func (u user) WebAuthnIcon() string {
	return ""
}

// WebAuthnCredentials helps implement the webauthn.User interface by loading
// the user's credentials from the underlying database.
func (u user) WebAuthnCredentials() []webauthn.Credential {
	var credential webauthn.Credential
	/*err := mapstructure.Decode(u.mfaData.Payload["default"], &credential)
	if err != nil {
		return nil
	}*/

	jsonBytes, err := json.Marshal(u.mfaData.Payload["default"])
	if err != nil {
		return nil
	}
	err = json.Unmarshal(jsonBytes, &credential)

	/*wcs := make([]webauthn.Credential, len(credentials))
	for i, cred := range credentials {
		credentialID, _ := base64.URLEncoding.DecodeString(cred.ID)
		wcs[i] = webauthn.Credential{
			ID:            credentialID,
			PublicKey:     cred.PublicKey,
			Authenticator: cred.WebauthnAuthenticator(),
		}
	}*/
	return []webauthn.Credential{credential}
}
