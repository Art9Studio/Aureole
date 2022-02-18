package fido2

import "aureole/internal/configs"

const (
	startRegistrationURL  = "/2fa/fido2/register"
	finishRegistrationURL = "/2fa/fido2/register/finish"
)

type config struct {
	AttestationType   string `mapstructure:"attestation_type"`
	AuthenticatorType string `mapstructure:"authenticator_type"`
}

func (c *config) setDefaults() {
	configs.SetDefault(&c.AttestationType, "direct")
}
