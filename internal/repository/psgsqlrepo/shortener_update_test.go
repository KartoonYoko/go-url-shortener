package psgsqlrepo

import (
	"context"
	"fmt"
	"sync"

	model "github.com/KartoonYoko/go-url-shortener/internal/model/shortener"
	"github.com/stretchr/testify/require"
)

// Test_psgsqlRepo_UpdateURLsDeletedFlag тестирует SQL запрос на обновление флага удаления URL
func (ts *PostgresTestSuite) Test_psgsqlRepo_UpdateURLsDeletedFlag() {
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

	urlsIDs := make([]string, 0, len(response))
	for _, v := range response {
		urlsIDs = append(urlsIDs, v.ShortURL)
	}

	modelToUpdateCh := fanIn(ctx, fanOut(ctx, generator(ctx, urlsIDs), 10)...)
	err = ts.psgsqlRepo.UpdateURLsDeletedFlag(ctx, userID, modelToUpdateCh)
	require.NoError(ts.T(), err)

	userURLS, err := ts.psgsqlRepo.GetUserURLs(ctx, userID)
	require.NoError(ts.T(), err)
	require.Equal(ts.T(), len(batch), len(userURLS))

	for _, v := range userURLS {
		found := false
		for _, b := range batch {
			if v.OriginalURL == b.OriginalURL {
				found = true
				break
			}
		}

		require.Equal(ts.T(), true, found)
	}
}

func generator(ctx context.Context, input []string) chan string {
	inputCh := make(chan string)

	go func() {
		defer close(inputCh)

		for _, data := range input {
			select {
			case <-ctx.Done():
				return
			case inputCh <- data:
			}
		}
	}()

	return inputCh
}

// createModelToUpdateFlag создаёт модель обновления флага из ID'шника URL'а
func createModelToUpdateFlag(ctx context.Context, inputCh chan string) chan model.UpdateURLDeletedFlag {
	result := make(chan model.UpdateURLDeletedFlag)

	go func() {
		defer close(result)

		for data := range inputCh {
			m := model.UpdateURLDeletedFlag{
				URLID: data,
			}

			select {
			case <-ctx.Done():
				return
			case result <- m:
			}
		}
	}()
	return result
}

// fanOut принимает канал данных, порождает numWorkers горутин
func fanOut(ctx context.Context, inputCh chan string, numWorkers int) []chan model.UpdateURLDeletedFlag {
	// каналы, в которые отправляются результаты
	channels := make([]chan model.UpdateURLDeletedFlag, numWorkers)

	for i := 0; i < numWorkers; i++ {
		channels[i] = createModelToUpdateFlag(ctx, inputCh)
	}

	// возвращаем слайс каналов
	return channels
}

// fanIn объединяет несколько каналов resultChs в один.
func fanIn(ctx context.Context, resultChs ...chan model.UpdateURLDeletedFlag) chan model.UpdateURLDeletedFlag {
	// конечный выходной канал в который отправляем данные из всех каналов из слайса, назовём его результирующим
	finalCh := make(chan model.UpdateURLDeletedFlag)
	var wg sync.WaitGroup
	for _, ch := range resultChs {
		chClosure := ch

		wg.Add(1)

		go func() {
			defer wg.Done()

			for data := range chClosure {
				select {
				case <-ctx.Done():
					return
				case finalCh <- data:
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(finalCh)
	}()

	// возвращаем результирующий канал
	return finalCh
}
