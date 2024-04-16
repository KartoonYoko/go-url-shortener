package psgsqlrepo

import (
	"context"

	"github.com/stretchr/testify/require"
)

// Test_psgsqlRepo_Ping тестирует пинг
func (ts *PostgresTestSuite) Test_psgsqlRepo_Ping() {
	ctx := context.Background()

	err := ts.psgsqlRepo.Ping(ctx)
	require.NoError(ts.T(), err)
}
