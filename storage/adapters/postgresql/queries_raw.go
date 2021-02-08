package postgresql

import "gouth/storage"

// Exec executes the given sql query with no returning results
func (s *Session) RawExec(sql string, args ...interface{}) error {
	_, err := s.conn.Exec(s.ctx, sql, args...)
	return err
}

// RawQuery executes the given sql query and returns results
func (s *Session) RawQuery(sql string, args ...interface{}) (storage.JSONCollResult, error) {
	var res interface{}

	err := s.conn.QueryRow(s.ctx, sql, args...).Scan(&res)
	if err != nil {
		return nil, err
	}

	return res, nil
}
