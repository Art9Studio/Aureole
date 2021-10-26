package postgresql

import (
	"aureole/internal/configs"
	"aureole/internal/plugins/storage/types"
	"context"
	"github.com/jackc/pgx/v4"
	"github.com/mitchellh/mapstructure"
)

// Storage represents a postgresql database
type Storage struct {
	rawConf *configs.Storage
	conf    *config
	conn    *pgx.Conn
	// for abstract queries
	relInfo map[types.CollPair]types.RelInfo
}

func (s *Storage) Init() error {
	adapterConf := &config{}
	if err := mapstructure.Decode(s.rawConf.Config, adapterConf); err != nil {
		return err
	}
	s.conf = adapterConf

	return s.Open()
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

	s.conn, err = pgx.Connect(context.Background(), url)
	return err
}

// Close terminates the currently active connection to the DBMS
func (s *Storage) Close() error {
	return s.conn.Close(context.Background())
}

func (s *Storage) Read() ([]byte, error) {
	panic("implement me")
}

func (s *Storage) Write(value []byte) error {
	panic("implement me")
}
