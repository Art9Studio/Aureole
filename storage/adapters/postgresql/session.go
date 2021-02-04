package postgresql

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v4"
	adapters "gouth/storage"
)

// Session represents a postgresql database
type Session struct {
	ctx      context.Context
	conn     *pgx.Conn
	connConf adapters.ConnectionConfig
	// for abstract queries
	relInfo map[adapters.CollectionPair]adapters.RelInfo
}

// Open creates connection with postgresql database
func (s *Session) Open() error {
	str, err := s.connConf.String()
	if err != nil {
		return err
	}

	config, err := pgx.ParseConfig(str)
	if err != nil {
		return err
	}

	conn, err := pgx.ConnectConfig(s.ctx, config)
	if err != nil {
		return err
	}

	s.conn = conn
	return nil
}

// ConnectionConfig returns the connection url that was used to set up the adapter
func (s *Session) ConnConfig() adapters.ConnectionConfig {
	return s.connConf
}

// Ping returns an error if the DBMS could not be reached
func (s *Session) Ping() error {
	var o int
	err := s.conn.QueryRow(context.Background(), "select 1").Scan(&o)
	if err != nil {
		return err
	}

	if o != 1 {
		return errors.New("got invalid data")
	}
	return nil
}

// Close terminates the currently active connection to the DBMS
func (s *Session) Close() error {
	return s.conn.Close(s.ctx)
}
