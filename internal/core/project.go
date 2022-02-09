package core

import (
	"aureole/internal/configs"
	"aureole/internal/plugins"
	"github.com/jpillora/overseer"
	"log"
	"net"
	"net/url"
	"os"
)

func RunReloadableAureole() {
	overseer.Run(overseer.Config{
		Program: runReloadableAureole,
		Address: ":" + getAureolePort(),
	})
}

func runReloadableAureole(state overseer.State) {
	run(state.Listener)
}

func RunAureole() {
	port := getAureolePort()
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		panic(err)
	}
	run(ln)
}

func run(ln net.Listener) {
	conf, err := configs.LoadMainConfig()
	if err != nil {
		log.Panic(err)
	}
	Init(conf)

	err = p.runServer(ln)
	if err != nil {
		log.Panic(err)
	}
}

func getAureolePort() (port string) {
	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = "3000"
	}
	return port
}

type project struct {
	apiVersion string
	testRun    bool
	pingPath   string
	apps       map[string]*app
	router     *router
}

func (p *project) runServer(ln net.Listener) error {
	server := createServer(p.router)
	return server.Listener(ln)
}

type (
	app struct {
		name           string
		url            *url.URL
		pathPrefix     string
		authSessionExp int
		service        service
		authenticators map[string]plugins.Authenticator
		authorizer     plugins.Authorizer
		secondFactors  map[string]plugins.SecondFactor
		idManager      plugins.IDManager
		storages       map[string]plugins.Storage
		cryptoStorages map[string]plugins.CryptoStorage
		senders        map[string]plugins.Sender
		cryptoKeys     map[string]plugins.CryptoKey
		admins         map[string]plugins.Admin
		ui             plugins.UI
	}

	service struct {
		signKey plugins.CryptoKey
		encKey  plugins.CryptoKey
		storage plugins.Storage
	}
)

func (a *app) getServiceSignKey() (plugins.CryptoKey, bool) {
	if a.service.signKey == nil {
		return nil, false
	}
	return a.service.signKey, true
}

func (a *app) getServiceEncKey() (plugins.CryptoKey, bool) {
	if a.service.encKey == nil {
		return nil, false
	}
	return a.service.encKey, true
}

func (a *app) getServiceStorage() (plugins.Storage, bool) {
	if a.service.storage == nil {
		return nil, false
	}
	return a.service.storage, true
}

func (a *app) getIDManager() (plugins.IDManager, bool) {
	if a.idManager == nil {
		return nil, false
	}
	return a.idManager, true
}

func (a *app) getAuthorizer() (plugins.Authorizer, bool) {
	if a.authorizer == nil {
		return nil, false
	}
	return a.authorizer, true
}

func (a *app) getSecondFactors() (map[string]plugins.SecondFactor, bool) {
	if len(a.secondFactors) == 0 {
		return nil, false
	}
	return a.secondFactors, true
}

func (a *app) getStorage(name string) (plugins.Storage, bool) {
	storage, ok := a.storages[name]
	return storage, ok
}

func (a *app) getCryptoStorage(name string) (plugins.CryptoStorage, bool) {
	cryptoStorage, ok := a.cryptoStorages[name]
	return cryptoStorage, ok
}

func (a *app) getSender(name string) (plugins.Sender, bool) {
	sender, ok := a.senders[name]
	return sender, ok
}

func (a *app) getCryptoKey(name string) (plugins.CryptoKey, bool) {
	cryptoKey, ok := a.cryptoKeys[name]
	return cryptoKey, ok
}
