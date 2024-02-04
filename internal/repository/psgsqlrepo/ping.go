package psgsqlrepo

import "context"

func (s *psgsqlRepo) Ping(ctx context.Context) error {
	return s.conn.PingContext(ctx)
}
