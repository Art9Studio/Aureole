package main

import _ "aureole/plugins/storage/postgresql"
import _ "aureole/plugins/authn/pwbased"
import _ "aureole/plugins/pwhasher/argon2"
import _ "aureole/plugins/pwhasher/pbkdf2"
import _ "aureole/plugins/sender/email"
import _ "aureole/plugins/authz/session"
import _ "aureole/plugins/authz/jwt"
import _ "aureole/plugins/cryptokey/jwk"
import _ "aureole/plugins/authn/phonebased"
import _ "aureole/plugins/sender/twilio"
