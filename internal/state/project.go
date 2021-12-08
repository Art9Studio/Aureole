package state

import (
	adminTypes "aureole/internal/plugins/admin/types"
	cryptoKeyTypes "aureole/internal/plugins/cryptokey/types"
	pwhasherTypes "aureole/internal/plugins/pwhasher/types"
	senderTypes "aureole/internal/plugins/sender/types"
	storageTypes "aureole/internal/plugins/storage/types"
	"aureole/internal/state/app"
	"fmt"
)

type (
	Project struct {
		APIVersion string
		TestRun    bool
		PingPath   string
		Apps       map[string]*app.App
		Storages   map[string]storageTypes.Storage
		Hashers    map[string]pwhasherTypes.PwHasher
		Senders    map[string]senderTypes.Sender
		CryptoKeys map[string]cryptoKeyTypes.CryptoKey
		Admins     map[string]adminTypes.Admin
	}
)

func (p *Project) IsTestRun() bool {
	return p.TestRun
}

func (p *Project) GetApp(name string) (*app.App, error) {
	a, ok := p.Apps[name]
	if !ok {
		return nil, fmt.Errorf("can't find app named '%s'", name)
	}

	return a, nil
}

func (p *Project) GetStorage(name string) (storageTypes.Storage, error) {
	s, ok := p.Storages[name]
	if !ok || s == nil {
		return nil, fmt.Errorf("can't find storage named '%s'", name)
	}

	return s, nil
}

func (p *Project) GetHasher(name string) (pwhasherTypes.PwHasher, error) {
	h, ok := p.Hashers[name]
	if !ok || h == nil {
		return nil, fmt.Errorf("can't find hasher named '%s'", name)
	}

	return h, nil
}

func (p *Project) GetSender(name string) (senderTypes.Sender, error) {
	s, ok := p.Senders[name]
	if !ok || s == nil {
		return nil, fmt.Errorf("can't find sender named '%s'", name)
	}

	return s, nil
}

func (p *Project) GetCryptoKey(name string) (cryptoKeyTypes.CryptoKey, error) {
	k, ok := p.CryptoKeys[name]
	if !ok || k == nil {
		return nil, fmt.Errorf("can't find crypto key named '%s'", name)
	}

	return k, nil
}
