package psgsqlrepo

import "context"

// Ping реализует Pinger
func (s *psgsqlRepo) Ping(ctx context.Context) error {
	return s.conn.PingContext(ctx)
}
