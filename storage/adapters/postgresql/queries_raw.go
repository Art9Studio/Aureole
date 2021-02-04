package postgresql

import adapters "gouth/storage"

// Exec executes the given sql query with no returning results
func (s *Session) RawExec(sql string) error {
	_, err := s.conn.Exec(s.ctx, sql)
	return err
}

// RawQuery executes the given sql query and returns results
func (s *Session) RawQuery(sql string) (adapters.JSONCollectionResult, error) {
	var res interface{}

	err := s.conn.QueryRow(s.ctx, sql).Scan(&res)
	if err != nil {
		return nil, err
	}

	return res, nil
}
