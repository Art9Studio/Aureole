package postgresql

import (
	"aureole/configs"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_pgAdapter_Get(t *testing.T) {
	adapter := pgAdapter{}

	validConf := &configs.Storage{
		Config: configs.RawConfig{
			"user":     "root",
			"password": "password",
			"host":     "localhost",
			"port":     "5432",
			"database": "test",
			"options":  map[string]string{},
		},
	}
	invalidConf := &configs.Storage{
		Config: configs.RawConfig{
			"user":     "root",
			"password": "password",
			"host":     "localhost",
			"port":     "5432",
			"database": "test",
			"options":  map[string]string{},
		},
	}

	usersSess, err := adapter.Get(validConf)
	assert.NoError(t, err)
	assert.NotNil(t, usersSess)

	usersSess, err = adapter.Get(invalidConf)
	assert.Error(t, err)
	assert.Nil(t, usersSess)
}
