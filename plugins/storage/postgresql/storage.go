package postgresql

import (
	"aureole/configs"
	"aureole/internal/collections"
	"aureole/internal/plugins/storage"
	"aureole/internal/plugins/storage/types"
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/mitchellh/mapstructure"
	"time"
)

// Storage represents a postgresql database
type Storage struct {
	rawConf    *configs.Storage
	conf       *config
	conn       *pgx.Conn
	gcInterval time.Duration
	gcDone     chan struct{}
	// for abstract queries
	relInfo map[types.CollPair]types.RelInfo
}

func (s *Storage) Init() error {
	adapterConf := &config{}
	if err := mapstructure.Decode(s.rawConf.Config, adapterConf); err != nil {
		return err
	}
	s.conf = adapterConf
	s.gcDone = make(chan struct{})

	return s.Open()
}

func (s *Storage) CheckFeaturesAvailable(requiredFeatures []string) error {
	return storage.CheckFeaturesAvailable(requiredFeatures, AdapterFeatures)
}

// Open creates connection with postgresql database
func (s *Storage) Open() error {
	var url string
	var err error

	if s.conf.Url == "" {
		url, err = s.conf.ToURL()
		if err != nil {
			return err
		}
	} else {
		url = s.conf.Url
	}

	conn, err := pgx.Connect(context.Background(), url)
	if err != nil {
		return err
	}

	s.conn = conn
	return nil
}

// Close terminates the currently active connection to the DBMS
func (s *Storage) Close() error {
	s.gcDone <- struct{}{}
	return s.conn.Close(context.Background())
}

// IsCollExists checks whether the given collection exists
func (s *Storage) IsCollExists(spec collections.Specification) (bool, error) {
	// TODO: use current schema instead constant 'public'
	sql := fmt.Sprintf(
		"select exists (select from pg_tables where schemaname = 'public' AND tablename = '%s');",
		spec.Name)
	res, err := s.RawQuery(sql)
	if err != nil {
		return false, err
	}

	return res.(bool), nil
}
