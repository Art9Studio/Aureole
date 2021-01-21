package postgresql

import (
	"context"
	adapters "gouth/storage"
)

// AdapterName is the internal name of the adapter
const AdapterName = "postgresql"

// init initializes package by register adapter
func init() {
	adapters.RegisterAdapter(AdapterName, &pgAdapter{})
}

// pgAdapter represents adapter for postgresql database
type pgAdapter struct {
}

// OpenUrl attempts to establish a connection with a db by connection url
func (pg pgAdapter) OpenUrl(connUrl adapters.ConnectionString) (adapters.Session, error) {
	connConf, err := ParseUrl(connUrl.String())
	if err != nil {
		return nil, err
	}

	return pg.OpenConfig(connConf)
}

// OpenConfig attempts to establish a connection with a db by connection config
func (pg pgAdapter) OpenConfig(connConf adapters.ConnectionConfig) (adapters.Session, error) {
	sess := &Session{
		ctx:      context.Background(),
		connConf: connConf,
	}

	if err := sess.Open(); err != nil {
		return nil, err
	}

	return sess, nil
}
