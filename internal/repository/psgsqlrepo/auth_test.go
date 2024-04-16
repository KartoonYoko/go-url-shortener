package psgsqlrepo

import (
	"context"

	"github.com/stretchr/testify/require"
)

// Test_psgsqlRepo_GetNewUserID тестирует SQL запрос создания нового пользователя
func (ts *PostgresTestSuite) Test_psgsqlRepo_GetNewUserID() {
	ctx := context.Background()

	userID, err := ts.psgsqlRepo.GetNewUserID(ctx)
	require.NoError(ts.T(), err)

	query := `
		SELECT id
		FROM users
		WHERE id = $1
	`
	err = ts.psgsqlRepo.conn.QueryRowContext(ctx, query, userID).Err()
	require.NoError(ts.T(), err)
}
