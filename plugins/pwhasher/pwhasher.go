package pwhasher

import "aureole/configs"

// PwHasher is an interface that defined method for pwhasher implementation
type PwHasher interface {
	// Hash returns hashed data encoded by base64
	HashPw(string) (string, error)

	// Compare compares plain data and hashed data encoded by base64
	ComparePw(string, string) (bool, error)
}

// New returns desired PwHasher depends on the given configs
func New(algoName string, rawConf *configs.RawConfig) (PwHasher, error) {
	adapter, err := GetAdapter(algoName)
	if err != nil {
		return nil, err
	}

	return adapter.GetPwHasher(rawConf)
}
