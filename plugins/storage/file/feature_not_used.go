package file

import "aureole/internal/plugins/storage/types"

func (s Storage) NativeQuery(s2 string, i ...interface{}) (types.JSONCollResult, error) {
	panic("implement me")
}

func (s Storage) RelInfo() map[types.CollPair]types.RelInfo {
	panic("implement me")
}

func (s Storage) Ping() error {
	return nil
}

func (s Storage) RawExec(s2 string, i ...interface{}) error {
	panic("implement me")
}

func (s Storage) RawQuery(s2 string, i ...interface{}) (types.JSONCollResult, error) {
	panic("implement me")
}

func (s Storage) Close() error {
	panic("implement me")
}
