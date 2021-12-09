package argon2

import (
	"aureole/internal/configs"
	"aureole/internal/core"
	"aureole/internal/plugins"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"
	a2 "golang.org/x/crypto/argon2"
)

const pluginID = "7416"

// argon2 represents argon2 hasher
type argon2 struct {
	pluginApi core.PluginAPI
	rawConf   *configs.PwHasher
	conf      *config
}

func (a *argon2) Init(api core.PluginAPI) error {
	a.pluginApi = api
	adapterConf := &config{}
	if err := mapstructure.Decode(a.rawConf.Config, adapterConf); err != nil {
		return err
	}
	adapterConf.setDefaults()
	a.conf = adapterConf

	return nil
}

func (a *argon2) GetMetaData() plugins.Meta {
	return plugins.Meta{
		Type: adapterName,
		Name: a.rawConf.Name,
		ID:   pluginID,
	}
}

// HashPw returns a argon2 pwhasher of a plain-text password using the provided algorithm
// parameters. The returned pwhasher follows the format used by the argon2 reference
// C implementation and contains the base64-encoded argon2 derived key prefixed
// by the salt and parameters. It looks like this:
//
//		$argon2i$v=19$m=65536,t=3,p=2$c29tZXNhbHQ$RdescudvJCsgt3ub+b+dWRWJTmaaJObG
//
func (a *argon2) HashPw(pw string) (string, error) {
	salt := make([]byte, a.conf.SaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	var key []byte

	// todo: save chosen function in context when init and use it here
	switch a.conf.Kind {
	case "argon2i":
		key = a2.Key([]byte(pw), salt, a.conf.Iterations, a.conf.Memory, a.conf.Parallelism, a.conf.KeyLen)
	case "argon2id":
		key = a2.IDKey([]byte(pw), salt, a.conf.Iterations, a.conf.Memory, a.conf.Parallelism, a.conf.KeyLen)
	}

	hashed := fmt.Sprintf("$%s$v=%d$m=%d,t=%d,p=%d$%s$%s",
		a.conf.Kind,
		a2.Version,
		a.conf.Memory,
		a.conf.Iterations,
		a.conf.Parallelism,
		base64.StdEncoding.EncodeToString(salt),
		base64.StdEncoding.EncodeToString(key),
	)

	return hashed, nil
}

// ComparePw performs a constant-time comparison between a plain-text password and
// argon2 pwhasher, using the parameters and salt contained in the pwhasher
// It returns true if they match, otherwise it returns false
func (*argon2) ComparePw(pw, hash string) (bool, error) {
	conf, salt, key, err := decodePwHash(hash)
	if err != nil {
		return false, err
	}

	var otherKey []byte

	switch conf.Kind {
	case "argon2i":
		otherKey = a2.Key([]byte(pw), salt, conf.Iterations, conf.Memory, conf.Parallelism, conf.KeyLen)
	case "argon2id":
		otherKey = a2.IDKey([]byte(pw), salt, conf.Iterations, conf.Memory, conf.Parallelism, conf.KeyLen)
	}

	if subtle.ConstantTimeCompare(key, otherKey) == 1 {
		return true, nil
	}

	return false, nil
}

// decodePwHash expects a pwhasher created from this package, and parses it to return the config
// used to create it, as well as the salt and key
func decodePwHash(hash string) (*config, []byte, []byte, error) {
	vals := strings.Split(hash, "$")
	if len(vals) != 6 {
		return nil, nil, nil, errors.New("argon2: the encoded pwhasher is not in the correct format")
	}

	var v int
	_, err := fmt.Sscanf(vals[2], "v=%d", &v)
	if err != nil {
		return nil, nil, nil, err
	}

	if v != a2.Version {
		return nil, nil, nil, errors.New("argon2: incompatible version of argon2")
	}

	conf := &config{}
	conf.Kind = vals[1]
	_, err = fmt.Sscanf(vals[3], "m=%d,t=%d,p=%d", &conf.Memory, &conf.Iterations, &conf.Parallelism)
	if err != nil {
		return nil, nil, nil, err
	}

	salt, err := base64.StdEncoding.DecodeString(vals[4])
	if err != nil {
		return nil, nil, nil, err
	}
	conf.SaltLen = uint32(len(salt))

	key, err := base64.StdEncoding.DecodeString(vals[5])
	if err != nil {
		return nil, nil, nil, err
	}
	conf.KeyLen = uint32(len(key))

	return conf, salt, key, nil
}
