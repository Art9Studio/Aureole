package pem

import (
	"aureole/internal/configs"
	"aureole/internal/core"
	"aureole/internal/plugins"
	"encoding/json"
	"fmt"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/mitchellh/mapstructure"
	"net/http"
	"strings"
	"sync"
	"time"
)

const pluginID = "6374"

type pem struct {
	pluginApi       core.PluginAPI
	rawConf         *configs.CryptoKey
	conf            *config
	keyStorage      plugins.KeyStorage
	refreshDone     chan struct{}
	refreshInterval time.Duration
	muSet           sync.RWMutex
	privateSet      jwk.Set
	publicSet       jwk.Set
}

func (p *pem) Init(api core.PluginAPI) (err error) {
	p.pluginApi = api
	if p.conf, err = initConfig(&p.rawConf.Config); err != nil {
		return err
	}
	p.conf.PathPrefix = "/" + strings.ReplaceAll(p.rawConf.Name, "_", "-")

	p.keyStorage, err = p.pluginApi.GetKeyStorage(p.conf.Storage)
	if err != nil {
		return fmt.Errorf("key keyStorage named '%s' is not declared", p.conf.Storage)
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

func (p *pem) GetMetaData() plugins.Meta {
	return plugins.Meta{
		Type: adapterName,
		Name: p.rawConf.Name,
		ID:   pluginID,
	}
}

func (p *pem) GetPrivateSet() jwk.Set {
	p.muSet.RLock()
	privSet := p.privateSet
	p.muSet.RUnlock()
	return privSet
}

func (p *pem) GetPublicSet() jwk.Set {
	p.muSet.RLock()
	pubSet := p.publicSet
	p.muSet.RUnlock()
	return pubSet
}

func initConfig(rawConf *configs.RawConfig) (*config, error) {
	adapterConf := &config{}
	if err := mapstructure.Decode(rawConf, adapterConf); err != nil {
		return nil, err
	}

	return adapterConf, nil
}

func initKeySets(p *pem) (err error) {
	var (
		rawKeys []byte
		keySet  jwk.Set
	)

	ok, err := p.keyStorage.Read(&rawKeys)
	if err != nil {
		return err
	}

	if ok {
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
		if err := p.keyStorage.Write(b); err != nil {
			return err
		}
	}

	setType, err := getKeySetType(keySet)
	if err != nil {
		return err
	}

	if setType == plugins.Private {
		p.privateSet = keySet
		if p.publicSet, err = jwk.PublicSetOf(p.privateSet); err != nil {
			return err
		}
	} else {
		p.publicSet = keySet
	}

	return nil
}

func createRoutes(p *pem) {
	routes := []*core.Route{
		{
			Method:  http.MethodGet,
			Path:    p.conf.PathPrefix + "/jwk",
			Handler: getJwkKeys(p),
		},
		{
			Method:  http.MethodGet,
			Path:    p.conf.PathPrefix + "/pem",
			Handler: getPemKeys(p),
		},
	}
	p.pluginApi.AddProjectRoutes(routes)
}

func refreshKeys(p *pem) {
	ticker := time.NewTicker(p.refreshInterval)
	defer ticker.Stop()
	for {
		select {
		case <-p.refreshDone:
			return
		case <-ticker.C:
			var (
				rawKeys []byte
				keySet  jwk.Set
			)

			ok, err := p.keyStorage.Read(&rawKeys)
			if err != nil {
				fmt.Printf("pem '%s': an error occured while refreshing keys: %v", p.rawConf.Name, err)
			}

			if ok {
				keySet, err = jwk.Parse(rawKeys)
				if err != nil {
					fmt.Printf("pem '%s': an error occured while refreshing keys: %v", p.rawConf.Name, err)
				}

				setType, err := getKeySetType(keySet)
				if err != nil {
					fmt.Printf("pem '%s': an error occured while refreshing keys: %v", p.rawConf.Name, err)
				}

				if setType == plugins.Private {
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
