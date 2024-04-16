package psgsqlrepo

import (
	"context"
	"fmt"

	model "github.com/KartoonYoko/go-url-shortener/internal/model/shortener"
	"github.com/stretchr/testify/require"
)

// Test_psgsqlRepo_SaveURL тестирует SQL запрос на вставку URL'a
func (ts *PostgresTestSuite) Test_psgsqlRepo_SaveURL() {
	ctx := context.Background()

	someURL := "https://someurl.example.com"
	userID, err := ts.psgsqlRepo.GetNewUserID(ctx)
	require.NoError(ts.T(), err)
	urlID, err := ts.psgsqlRepo.SaveURL(ctx, someURL, userID)
	require.NoError(ts.T(), err)

	gotURL, err := ts.psgsqlRepo.GetURLByID(ctx, urlID)
	require.NoError(ts.T(), err)

	require.Equal(ts.T(), someURL, gotURL)
}

// Test_psgsqlRepo_SaveURLsBatch тестирует SQL запрос на вставку множества URL'ов
func (ts *PostgresTestSuite) Test_psgsqlRepo_SaveURLsBatch() {
	ctx := context.Background()

	batchLength := 10
	batch := make([]model.CreateShortenURLBatchItemRequest, 0, batchLength)
	for i := 0; i < batchLength; i++ {
		batch = append(batch, model.CreateShortenURLBatchItemRequest{
			CorrelationID: fmt.Sprintf("%d", i+1),
			OriginalURL:   fmt.Sprintf("https://someurl%d.example.com", i+1),
		})
	}

	userID, err := ts.psgsqlRepo.GetNewUserID(ctx)
	require.NoError(ts.T(), err)
	response, err := ts.psgsqlRepo.SaveURLsBatch(ctx, batch, userID)
	require.NoError(ts.T(), err)

	require.Len(ts.T(), response, len(batch))

	for _, b := range batch {
		found := false
		for _, v := range response {
			if v.CorrelationID == b.CorrelationID {
				found = true
				break
			}
		}

		if !found {
			require.Fail(
				ts.T(),
				"not found correlation id %s with url %s in response",
				b.CorrelationID,
				b.OriginalURL)
		}
	}
}
