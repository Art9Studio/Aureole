package main

import (
	_ "aureole/plugins/authn/phone"
	_ "aureole/plugins/authn/pwbased"

	_ "aureole/plugins/authn/email"

	_ "aureole/plugins/authn/google"

	_ "aureole/plugins/authn/vk"

	_ "aureole/plugins/authn/facebook"

	_ "aureole/plugins/authn/apple"

	_ "aureole/plugins/authz/jwt"

	_ "aureole/plugins/2fa/authenticator"

	_ "aureole/plugins/2fa/sms"

	_ "aureole/plugins/2fa/yubikey"

	_ "aureole/plugins/identity/jwt_webhook"

	_ "aureole/plugins/identity/standard"

	_ "aureole/plugins/kstorage/file"

	_ "aureole/plugins/kstorage/url"

	_ "aureole/plugins/kstorage/vault"

	_ "aureole/plugins/storage/etcd"

	_ "aureole/plugins/storage/redis"

	_ "aureole/plugins/storage/memory"

	_ "aureole/plugins/pwhasher/argon2"

	_ "aureole/plugins/pwhasher/pbkdf2"

	_ "aureole/plugins/cryptokey/jwk"

	_ "aureole/plugins/cryptokey/pem"

	_ "aureole/plugins/sender/email"

	_ "aureole/plugins/sender/twilio"

	_ "aureole/plugins/admin/urls"
)
