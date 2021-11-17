package state

import (
	adminT "aureole/internal/plugins/admin/types"
	cryptoKeyT "aureole/internal/plugins/cryptokey/types"
	kstorageT "aureole/internal/plugins/kstorage/types"
	pwhasherT "aureole/internal/plugins/pwhasher/types"
	senderT "aureole/internal/plugins/sender/types"
	storageT "aureole/internal/plugins/storage/types"
	"aureole/internal/state/app"
	"errors"
	"fmt"
)

type (
	Project struct {
		APIVersion  string
		TestRun     bool
		PingPath    string
		Service     service
		Apps        map[string]*app.App
		Storages    map[string]storageT.Storage
		KeyStorages map[string]kstorageT.KeyStorage
		Hashers     map[string]pwhasherT.PwHasher
		Senders     map[string]senderT.Sender
		CryptoKeys  map[string]cryptoKeyT.CryptoKey
		Admins      map[string]adminT.Admin
	}

	service struct {
		internalKey cryptoKeyT.CryptoKey
		storage     storageT.Storage
	}
)

func (p *Project) IsTestRun() bool {
	return p.TestRun
}

func (p *Project) GetServiceKey() (cryptoKeyT.CryptoKey, error) {
	if p.Service.internalKey == nil {
		return nil, errors.New("cannot find service key")
	}
	return p.Service.internalKey, nil
}

func (p *Project) GetServiceStorage() (storageT.Storage, error) {
	if p.Service.storage == nil {
		return nil, errors.New("cannot find service kstorage")
	}
	return p.Service.storage, nil
}

func (p *Project) GetApp(name string) (*app.App, error) {
	a, ok := p.Apps[name]
	if !ok {
		return nil, fmt.Errorf("can't find app named '%s'", name)
	}

	return a, nil
}

func (p *Project) GetStorage(name string) (storageT.Storage, error) {
	s, ok := p.Storages[name]
	if !ok || s == nil {
		return nil, fmt.Errorf("can't find storage named '%s'", name)
	}

	return s, nil
}

func (p *Project) GetKeyStorage(name string) (kstorageT.KeyStorage, error) {
	s, ok := p.KeyStorages[name]
	if !ok || s == nil {
		return nil, fmt.Errorf("can't find key storage named '%s'", name)
	}

	return s, nil
}

func (p *Project) GetHasher(name string) (pwhasherT.PwHasher, error) {
	h, ok := p.Hashers[name]
	if !ok || h == nil {
		return nil, fmt.Errorf("can't find hasher named '%s'", name)
	}

	return h, nil
}

func (p *Project) GetSender(name string) (senderT.Sender, error) {
	s, ok := p.Senders[name]
	if !ok || s == nil {
		return nil, fmt.Errorf("can't find sender named '%s'", name)
	}

	return s, nil
}

func (p *Project) GetCryptoKey(name string) (cryptoKeyT.CryptoKey, error) {
	k, ok := p.CryptoKeys[name]
	if !ok || k == nil {
		return nil, fmt.Errorf("can't find crypto key named '%s'", name)
	}

	return k, nil
}
