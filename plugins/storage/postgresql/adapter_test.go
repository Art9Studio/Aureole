package postgresql

import (
	"aureole/internal/configs"
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

	usersSess := adapter.Create(validConf)
	assert.NotNil(t, usersSess)
}
