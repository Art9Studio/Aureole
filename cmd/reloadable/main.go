package main

import (
	"aureole/internal/core"
	_ "aureole/plugins/2fa/authenticator"
	_ "aureole/plugins/2fa/sms"
	_ "aureole/plugins/2fa/yubikey"
	_ "aureole/plugins/admin/urls"
	_ "aureole/plugins/authn/apple"
	_ "aureole/plugins/authn/email"
	_ "aureole/plugins/authn/facebook"
	_ "aureole/plugins/authn/google"
	_ "aureole/plugins/authn/phone"
	_ "aureole/plugins/authn/pwbased"
	_ "aureole/plugins/authn/vk"
	_ "aureole/plugins/authz/jwt"
	_ "aureole/plugins/crypto-key/jwk"
	_ "aureole/plugins/crypto-key/pem"
	_ "aureole/plugins/crypto-storage/file"
	_ "aureole/plugins/crypto-storage/url"
	_ "aureole/plugins/crypto-storage/vault"
	_ "aureole/plugins/identity/jwt_webhook"
	_ "aureole/plugins/identity/standard"
	_ "aureole/plugins/sender/email"
	_ "aureole/plugins/sender/twilio"
	_ "aureole/plugins/storage/etcd"
	_ "aureole/plugins/storage/memory"
	_ "aureole/plugins/storage/redis"
	_ "aureole/plugins/ui/standard"
)

func main() {
	core.RunReloadableAureole()
}
