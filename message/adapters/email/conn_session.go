package email

import (
	"aureole/message"
)

type ConnSession struct {
	connConf message.ConnConfig
}

func (s *ConnSession) Open() error {
	s.conn = conn
	return nil
}

func (s *ConnSession) GetConfig() message.ConnConfig {
	return s.connConf
}

func (s *ConnSession) Close() error {
	return s.conn.Close(s.ctx)
}
