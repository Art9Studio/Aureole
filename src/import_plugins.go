package main

import (
	_ "aureole/plugins/auth/apple"
	_ "aureole/plugins/auth/email"
	_ "aureole/plugins/auth/facebook"
	_ "aureole/plugins/auth/google"
	_ "aureole/plugins/auth/phone"
	_ "aureole/plugins/auth/pwbased"
	_ "aureole/plugins/auth/vk"
	//
	_ "aureole/plugins/crypto-key/jwk"
	//_ "aureole/plugins/crypto-key/pem"
	_ "aureole/plugins/issuer/jwt"
	//_ "aureole/plugins/mfa/otp"
	// _ "aureole/plugins/mfa/sms"
	//_ "aureole/plugins/mfa/yubikey"
	// _ "aureole/plugins/identity/jwt_webhook"
	// _ "aureole/plugins/identity/standard"
	_ "aureole/plugins/crypto-storage/file"
	_ "aureole/plugins/crypto-storage/url"
	_ "aureole/plugins/crypto-storage/vault"
	_ "aureole/plugins/root/urls"
	//_ "aureole/plugins/sender/email"
	//_ "aureole/plugins/sender/twilio"
	//_ "aureole/plugins/storage/etcd"
	_ "aureole/plugins/storage/memory"
	//_ "aureole/plugins/storage/redis"
)
