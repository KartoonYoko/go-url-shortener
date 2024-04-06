package psgsqlrepo

import (
	"context"

	"github.com/stretchr/testify/require"
)

// Test_psgsqlRepo_GetURLByID проверяет SQL запрос на вставку URL'a
func (ts *PostgresTestSuite) Test_psgsqlRepo_GetURLByID() {
	ctx := context.Background()

	someURL := "https://someurl.example.com"
	userID, err := ts.psgsqlRepo.GetNewUserID(ctx)
	require.NoError(ts.T(), err)
	urlID, err := ts.psgsqlRepo.SaveURL(ctx, someURL, userID)
	require.NoError(ts.T(), err)

	checkableURL, err := ts.psgsqlRepo.GetURLByID(ctx, urlID)
	require.NoError(ts.T(), err)

	require.Equal(ts.T(), someURL, checkableURL)
}

// Test_psgsqlRepo_GetUserURLs проверяет SQL запрос на получение URL'ов конкретным пользователем
func (ts *PostgresTestSuite) Test_psgsqlRepo_GetUserURLs() {
	ctx := context.Background()

	urls := []string{
		"https://someurl.example.com",
		"https://someurl.example1.com",
		"https://someurl.example2.com",
		"https://someurl.example3.com",
	}
	userID, err := ts.psgsqlRepo.GetNewUserID(ctx)
	require.NoError(ts.T(), err)

	// создадим для пользователя URL'ы
	m := map[string]string{}
	for _, u := range urls {
		urlID, err := ts.psgsqlRepo.SaveURL(ctx, u, userID)
		require.NoError(ts.T(), err)
		m[u] = urlID
	}
	// проверим что все URL'ы отработаны
	for _, v := range urls {
		_, ok := m[v]
		require.Equal(ts.T(), true, ok)
	}

	// проверим, что основной метод возвращает все добавленные ранее URL'ы
	res, err := ts.psgsqlRepo.GetUserURLs(ctx, userID)
	require.NoError(ts.T(), err)
	require.Equal(ts.T(), len(urls), len(res), "Length of added urls and got urls are not equal")

	for _, v := range res {
		_, ok := m[v.OriginalURL]
		require.Equal(ts.T(), true, ok)
	}
}
