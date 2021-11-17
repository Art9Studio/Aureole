package vault

import (
	"aureole/internal/configs"
	"encoding/json"
	"github.com/hashicorp/vault/api"
	"github.com/mitchellh/mapstructure"
)

type Storage struct {
	rawConf *configs.KeyStorage
	conf    *config
	client  *api.Client
}

func (s *Storage) Init() error {
	adapterConf := &config{}
	if err := mapstructure.Decode(s.rawConf.Config, adapterConf); err != nil {
		return err
	}
	s.conf = adapterConf

	client, err := api.NewClient(&api.Config{Address: s.conf.Address})
	if err != nil {
		return err
	}
	client.SetToken(s.conf.Token)
	s.client = client

	return nil
}

func (s *Storage) Write(v []byte) error {
	_, err := s.client.Logical().WriteBytes(s.conf.Path, v)
	return err
}

func (s *Storage) Read(v *[]byte) (ok bool, err error) {
	scr, err := s.client.Logical().Read(s.conf.Path)
	if err != nil {
		return false, err
	} else if scr == nil {
		return false, nil
	}

	*v, err = json.Marshal(scr.Data)
	if err != nil {
		return false, err
	} else if string(*v) == "null" {
		return false, nil
	}
	return true, nil
}
