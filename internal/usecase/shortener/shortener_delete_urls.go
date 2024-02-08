package shortener

import (
	"context"
	"sync"

	model "github.com/KartoonYoko/go-url-shortener/internal/model/shortener"
)

func (s *shortenerUsecase) DeleteURLs(ctx context.Context, userID string, urlsIDs []string) error {
	modelToUpdateCh := fanIn(ctx, fanOut(ctx, generator(ctx, urlsIDs), 10)...)
	return s.repository.UpdateURLsDeletedFlag(ctx, userID, modelToUpdateCh)
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
