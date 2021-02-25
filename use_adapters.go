package main

import _ "aureole/plugins/storage/adapters/postgresql"
import _ "aureole/plugins/authn/adapters/pwbased"
import _ "aureole/plugins/pwhasher/adapters/argon2"
import _ "aureole/plugins/pwhasher/adapters/pbkdf2"
