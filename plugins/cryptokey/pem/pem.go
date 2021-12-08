package pem

import (
	"aureole/internal/configs"
	"aureole/internal/plugins/cryptokey"
	"aureole/internal/plugins/cryptokey/types"
	storageT "aureole/internal/plugins/storage/types"
	_interface "aureole/internal/router/interface"
	"encoding/json"
	"fmt"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/mitchellh/mapstructure"
	"strings"
	"sync"
	"time"
)

type Pem struct {
	rawConf         *configs.CryptoKey
	conf            *config
	storage         storageT.Storage
	refreshDone     chan struct{}
	refreshInterval time.Duration
	muSet           sync.RWMutex
	privateSet      jwk.Set
	publicSet       jwk.Set
}

func (p *Pem) Init() (err error) {
	p.rawConf.PathPrefix = "/" + strings.ReplaceAll(p.rawConf.Name, "_", "-")
	if p.conf, err = initConfig(&p.rawConf.Config); err != nil {
		return err
	}

	pluginApi := cryptokey.Repository.PluginApi
	p.storage, err = pluginApi.Project.GetStorage(p.conf.Storage)
	if err != nil {
		return fmt.Errorf("storage named '%s' is not declared\n", p.conf.Storage)
	}

	err = initKeySets(p)
	if err != nil {
		return err
	}
	createRoutes(p)

	if p.conf.RefreshInterval != 0 {
		p.refreshInterval = time.Duration(p.conf.RefreshInterval) * time.Second
		p.refreshDone = make(chan struct{})
		go refreshKeys(p)
	}

	return nil
}

func initConfig(rawConf *configs.RawConfig) (*config, error) {
	adapterConf := &config{}
	if err := mapstructure.Decode(rawConf, adapterConf); err != nil {
		return nil, err
	}

	return adapterConf, nil
}

func initKeySets(p *Pem) (err error) {
	var keySet jwk.Set

	rawKeys, err := p.storage.Read()
	if err != nil {
		return err
	}

	if len(rawKeys) != 0 {
		keySet, err = jwk.Parse(rawKeys, jwk.WithPEM(true))
		if err != nil {
			return err
		}
		if err := setAttr(keySet, p.conf.Alg); err != nil {
			return err
		}
	} else {
		keySet, err = generateKey()
		if err != nil {
			return err
		}

		b, err := json.MarshalIndent(keySet, "", "  ")
		if err != nil {
			return err
		}
		if err := p.storage.Write(b); err != nil {
			return err
		}
	}

	setType, err := getKeySetType(keySet)
	if err != nil {
		return err
	}

	if setType == types.Private {
		p.privateSet = keySet
		if p.publicSet, err = jwk.PublicSetOf(p.privateSet); err != nil {
			return err
		}
	} else {
		p.publicSet = keySet
	}

	return nil
}

func createRoutes(p *Pem) {
	routes := []*_interface.Route{
		{
			Method:  "GET",
			Path:    p.rawConf.PathPrefix + "/jwk",
			Handler: GetJwkKeys(p),
		},
		{
			Method:  "GET",
			Path:    p.rawConf.PathPrefix + "/pem",
			Handler: GetPemKeys(p),
		},
	}
	cryptokey.Repository.PluginApi.Router.AddProjectRoutes(routes)
}

func refreshKeys(p *Pem) {
	ticker := time.NewTicker(p.refreshInterval)
	defer ticker.Stop()
	for {
		select {
		case <-p.refreshDone:
			return
		case <-ticker.C:
			var keySet jwk.Set

			rawKeys, err := p.storage.Read()
			if err != nil {
				fmt.Printf("pem '%s': an error occured while refreshing keys: %v", p.rawConf.Name, err)
			}

			if len(rawKeys) != 0 {
				keySet, err = jwk.Parse(rawKeys)
				if err != nil {
					fmt.Printf("pem '%s': an error occured while refreshing keys: %v", p.rawConf.Name, err)
				}

				setType, err := getKeySetType(keySet)
				if err != nil {
					fmt.Printf("pem '%s': an error occured while refreshing keys: %v", p.rawConf.Name, err)
				}

				if setType == types.Private {
					pubSet, err := jwk.PublicSetOf(keySet)
					if err != nil {
						fmt.Printf("pem '%s': an error occured while refreshing keys: %v", p.rawConf.Name, err)
					}

					p.muSet.Lock()
					p.privateSet = keySet
					p.publicSet = pubSet
					p.muSet.Unlock()
				} else {
					p.muSet.Lock()
					p.publicSet = keySet
					p.muSet.Unlock()
				}

			} else {
				fmt.Printf("pem '%s': an error occured while refreshing keys: key is empty", p.rawConf.Name)
			}
		}
	}
}

func (p *Pem) GetPrivateSet() jwk.Set {
	p.muSet.RLock()
	privSet := p.privateSet
	p.muSet.RUnlock()
	return privSet
}

func (p *Pem) GetPublicSet() jwk.Set {
	p.muSet.RLock()
	pubSet := p.publicSet
	p.muSet.RUnlock()
	return pubSet
}
