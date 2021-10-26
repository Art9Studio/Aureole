package vault

import (
	"aureole/internal/configs"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/vault/api"
	"github.com/mitchellh/mapstructure"
)

type Storage struct {
	rawConf *configs.Storage
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

func (s *Storage) Read() ([]byte, error) {
	scr, err := s.client.Logical().Read(s.conf.Path)
	if err != nil {
		return nil, fmt.Errorf("read: %w", err)
	}
	if scr == nil {
		return []byte{}, nil
	}

	bytes, err := json.Marshal(scr.Data)
	if err != nil {
		return nil, err
	}

	if string(bytes) == "null" {
		return []byte{}, nil
	}

	return bytes, nil
}

func (s *Storage) Write(value []byte) error {
	_, err := s.client.Logical().WriteBytes(s.conf.Path, value)
	if err != nil {
		return fmt.Errorf("write: %w", err)
	}

	return nil
}
