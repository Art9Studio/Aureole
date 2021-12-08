package jwk

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

type (
	Jwk struct {
		rawConf         *configs.CryptoKey
		conf            *config
		storage         storageT.Storage
		refreshDone     chan struct{}
		refreshInterval time.Duration
		muSet           sync.RWMutex
		privateSet      jwk.Set
		publicSet       jwk.Set
	}
)

func (j *Jwk) Init() (err error) {
	j.rawConf.PathPrefix = "/" + strings.ReplaceAll(j.rawConf.Name, "_", "-")
	if j.conf, err = initConfig(&j.rawConf.Config); err != nil {
		return err
	}

	pluginApi := cryptokey.Repository.PluginApi
	j.storage, err = pluginApi.Project.GetStorage(j.conf.Storage)
	if err != nil {
		return fmt.Errorf("storage named '%s' is not declared\n", j.conf.Storage)
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

func initConfig(rawConf *configs.RawConfig) (*config, error) {
	adapterConf := &config{}
	if err := mapstructure.Decode(rawConf, adapterConf); err != nil {
		return nil, err
	}
	adapterConf.setDefaults()

	return adapterConf, nil
}

func initKeySets(j *Jwk) (err error) {
	var keySet jwk.Set

	rawKeys, err := j.storage.Read()
	if err != nil {
		return err
	}

	if len(rawKeys) != 0 {
		keySet, err = jwk.Parse(rawKeys)
		if err != nil {
			return err
		}
	} else {
		keySet, err = generateKey(j.conf)
		if err != nil {
			return err
		}

		b, err := json.MarshalIndent(keySet, "", "  ")
		if err != nil {
			return err
		}
		if err := j.storage.Write(b); err != nil {
			return err
		}
	}

	setType, err := getKeySetType(keySet)
	if err != nil {
		return err
	}

	if setType == types.Private {
		j.privateSet = keySet
		if j.publicSet, err = jwk.PublicSetOf(j.privateSet); err != nil {
			return err
		}
	} else {
		j.publicSet = keySet
	}

	return nil
}

func createRoutes(j *Jwk) {
	routes := []*_interface.Route{
		{
			Method:  "GET",
			Path:    j.rawConf.PathPrefix + "/jwk",
			Handler: GetJwkKeys(j),
		},
		{
			Method:  "GET",
			Path:    j.rawConf.PathPrefix + "/pem",
			Handler: GetPemKeys(j),
		},
	}
	cryptokey.Repository.PluginApi.Router.AddProjectRoutes(routes)
}

func refreshKeys(j *Jwk) {
	ticker := time.NewTicker(j.refreshInterval)
	defer ticker.Stop()
	for {
		select {
		case <-j.refreshDone:
			return
		case <-ticker.C:
			var keySet jwk.Set

			rawKeys, err := j.storage.Read()
			if err != nil {
				fmt.Printf("jwk '%s': an error occured while refreshing keys: %v", j.rawConf.Name, err)
			}

			if len(rawKeys) != 0 {
				keySet, err = jwk.Parse(rawKeys)
				if err != nil {
					fmt.Printf("jwk '%s': an error occured while refreshing keys: %v", j.rawConf.Name, err)
				}

				setType, err := getKeySetType(keySet)
				if err != nil {
					fmt.Printf("jwk '%s': an error occured while refreshing keys: %v", j.rawConf.Name, err)
				}

				if setType == types.Private {
					pubSet, err := jwk.PublicSetOf(keySet)
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
				fmt.Printf("jwk '%s': an error occured while refreshing keys: key is empty", j.rawConf.Name)
			}
		}
	}
}

func (j *Jwk) GetPrivateSet() jwk.Set {
	j.muSet.RLock()
	privSet := j.privateSet
	j.muSet.RUnlock()
	return privSet
}

func (j *Jwk) GetPublicSet() jwk.Set {
	j.muSet.RLock()
	pubSet := j.publicSet
	j.muSet.RUnlock()
	return pubSet
}
