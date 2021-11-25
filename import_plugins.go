package main

import _ "aureole/plugins/authn/pwbased"
import _ "aureole/plugins/authn/phone"
import _ "aureole/plugins/authn/email"
import _ "aureole/plugins/authn/google"
import _ "aureole/plugins/authn/vk"
import _ "aureole/plugins/authn/facebook"
import _ "aureole/plugins/authn/apple"

import _ "aureole/plugins/authz/jwt"

import _ "aureole/plugins/2fa/authenticator"
import _ "aureole/plugins/2fa/sms"
import _ "aureole/plugins/2fa/yubikey"

import _ "aureole/plugins/kstorage/file"
import _ "aureole/plugins/kstorage/url"
import _ "aureole/plugins/kstorage/vault"

import _ "aureole/plugins/storage/etcd"
import _ "aureole/plugins/storage/redis"
import _ "aureole/plugins/storage/memory"

import _ "aureole/plugins/pwhasher/argon2"
import _ "aureole/plugins/pwhasher/pbkdf2"

import _ "aureole/plugins/cryptokey/jwk"
import _ "aureole/plugins/cryptokey/pem"

import _ "aureole/plugins/sender/email"
import _ "aureole/plugins/sender/twilio"

import _ "aureole/plugins/admin/urls"
