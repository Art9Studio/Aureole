package jwk

import (
	"aureole/internal/configs"
	"aureole/internal/core"
	"aureole/internal/plugins"
	"encoding/json"
	"fmt"
	jwx "github.com/lestrrat-go/jwx/jwk"
	"github.com/mitchellh/mapstructure"
	"net/http"
	"strings"
	"sync"
	"time"
)

const pluginID = "7851"

type (
	jwk struct {
		pluginApi       core.PluginAPI
		rawConf         *configs.CryptoKey
		conf            *config
		keyStorage      plugins.KeyStorage
		refreshInterval time.Duration
		muSet           sync.RWMutex
		privateSet      jwx.Set
		publicSet       jwx.Set
		refreshDone     chan struct{}
	}
)

func (j *jwk) Init(api core.PluginAPI) (err error) {
	j.pluginApi = api
	if j.conf, err = initConfig(&j.rawConf.Config); err != nil {
		return err
	}
	j.conf.PathPrefix = "/" + strings.ReplaceAll(j.rawConf.Name, "_", "-")

	j.keyStorage, err = j.pluginApi.GetKeyStorage(j.conf.Storage)
	if err != nil {
		return fmt.Errorf("key storage named '%s' is not declared", j.conf.Storage)
	}

	err = initKeySets(j)
	if err != nil {
		return err
	}
	createRoutes(j)

	if j.conf.RefreshInterval != 0 {
		j.refreshInterval = time.Duration(j.conf.RefreshInterval) * time.Second
		j.refreshDone = make(chan struct{})
		go refreshKeys(j)
	}

	return nil
}

func (j *jwk) GetMetaData() plugins.Meta {
	return plugins.Meta{
		Type: adapterName,
		Name: j.rawConf.Name,
		ID:   pluginID,
	}
}

func initConfig(rawConf *configs.RawConfig) (*config, error) {
	adapterConf := &config{}
	if err := mapstructure.Decode(rawConf, adapterConf); err != nil {
		return nil, err
	}
	adapterConf.setDefaults()

	return adapterConf, nil
}

func (j *jwk) GetPrivateSet() jwx.Set {
	j.muSet.RLock()
	privSet := j.privateSet
	j.muSet.RUnlock()
	return privSet
}

func (j *jwk) GetPublicSet() jwx.Set {
	j.muSet.RLock()
	pubSet := j.publicSet
	j.muSet.RUnlock()
	return pubSet
}

func initKeySets(j *jwk) (err error) {
	var (
		rawKeys []byte
		keySet  jwx.Set
	)

	found, err := j.keyStorage.Read(&rawKeys)
	if err != nil {
		return err
	}

	if found {
		keySet, err = jwx.Parse(rawKeys)
		if err != nil {
			return err
		}
	} else {
		keySet, err = generateKey(j.conf)
		if err != nil {
			return err
		}

		b, err := json.Marshal(keySet)
		if err != nil {
			return err
		}
		if err := j.keyStorage.Write(b); err != nil {
			return err
		}
	}

	setType, err := getKeySetType(keySet)
	if err != nil {
		return err
	}

	if setType == plugins.Private {
		j.privateSet = keySet
		if j.publicSet, err = jwx.PublicSetOf(j.privateSet); err != nil {
			return err
		}
	} else {
		j.publicSet = keySet
	}

	return nil
}

func createRoutes(j *jwk) {
	routes := []*core.Route{
		{
			Method:  http.MethodGet,
			Path:    j.conf.PathPrefix + "/jwk",
			Handler: getJwkKeys(j),
		},
		{
			Method:  http.MethodGet,
			Path:    j.conf.PathPrefix + "/pem",
			Handler: getPemKeys(j),
		},
	}
	j.pluginApi.AddProjectRoutes(routes)
}

func refreshKeys(j *jwk) {
	ticker := time.NewTicker(j.refreshInterval)
	defer ticker.Stop()
	for {
		select {
		case <-j.refreshDone:
			return
		case <-ticker.C:
			var (
				rawKeys []byte
				keySet  jwx.Set
			)

			ok, err := j.keyStorage.Read(&rawKeys)
			if err != nil {
				fmt.Printf("jwk '%s': an error occured while refreshing keys: %v", j.rawConf.Name, err)
			}

			if ok {
				keySet, err = jwx.Parse(rawKeys)
				if err != nil {
					fmt.Printf("jwk '%s': an error occured while refreshing keys: %v", j.rawConf.Name, err)
				}

				setType, err := getKeySetType(keySet)
				if err != nil {
					fmt.Printf("jwk '%s': an error occured while refreshing keys: %v", j.rawConf.Name, err)
				}

				if setType == plugins.Private {
					pubSet, err := jwx.PublicSetOf(keySet)
					if err != nil {
						fmt.Printf("jwk '%s': an error occured while refreshing keys: %v", j.rawConf.Name, err)
					}

					j.muSet.Lock()
					j.privateSet = keySet
					j.publicSet = pubSet
					j.muSet.Unlock()
				} else {
					j.muSet.Lock()
					j.publicSet = keySet
					j.muSet.Unlock()
				}

			} else {
				fmt.Printf("jwk '%s': an error occured while refreshing keys: key not ok", j.rawConf.Name)
			}
		}
	}
}
