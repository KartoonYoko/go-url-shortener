package psgsqlrepo

import (
	"context"

	"github.com/stretchr/testify/require"
)

func (ts *PostgresTestSuite) Test_psgsqlRepo_GetStats() {
	ctx := context.Background()

	r, err := ts.psgsqlRepo.GetStats(ctx)
	require.NoError(ts.T(), err)
	require.NotNil(ts.T(), r)
}
