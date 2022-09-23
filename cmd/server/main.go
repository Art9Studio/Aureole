package main

import (
	_ "aureole/plugins/auth/apple"
	_ "aureole/plugins/auth/email"
	_ "aureole/plugins/auth/facebook"
	_ "aureole/plugins/auth/google"
	_ "aureole/plugins/auth/phone"
	_ "aureole/plugins/auth/pwbased"
	_ "aureole/plugins/auth/vk"
	_ "aureole/plugins/crypto-key/jwk"
	_ "aureole/plugins/crypto-key/pem"
	_ "aureole/plugins/crypto-storage/file"
	_ "aureole/plugins/crypto-storage/url"
	_ "aureole/plugins/crypto-storage/vault"
	//_ "aureole/plugins/identity-manager/jwt_webhook"
	_ "aureole/plugins/identity-manager/standard"
	_ "aureole/plugins/issuer/jwt"
	_ "aureole/plugins/mfa/otp"
	_ "aureole/plugins/mfa/sms"
	//_ "aureole/plugins/mfa/yubikey"
	_ "aureole/plugins/root/urls"
	_ "aureole/plugins/sender/email"
	_ "aureole/plugins/sender/twilio"
	//_ "aureole/plugins/storage/etcd"
	_ "aureole/plugins/storage/memory"
	_ "aureole/plugins/storage/redis"
)

import (
	"aureole/internal/core"
	"log"

	"aureole/internal/configs"
)

func main() {
	conf, err := configs.LoadMainConfig()
	if err != nil {
		log.Panic(err)
	}

	router := core.CreateRouter()
	project := core.InitProject(conf, router)
	err = core.RunServer(project, router)

	if err != nil {
		log.Panic(err)
	}
}
