package main

import (
	_ "aureole/plugins/authn/pwbased"
	_ "aureole/plugins/storage/postgresql"

	_ "aureole/plugins/pwhasher/argon2"

	_ "aureole/plugins/pwhasher/pbkdf2"

	_ "aureole/plugins/sender/email"

	_ "aureole/plugins/authz/session"

	_ "aureole/plugins/authz/jwt"

	_ "aureole/plugins/cryptokey/jwk"

	_ "aureole/plugins/authn/phone"

	_ "aureole/plugins/sender/twilio"

	_ "aureole/plugins/authn/email"
)
