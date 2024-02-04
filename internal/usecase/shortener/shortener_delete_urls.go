package shortener

import (
	"context"
	"sync"

	model "github.com/KartoonYoko/go-url-shortener/internal/model/shortener"
)

func (s *shortenerUsecase) DeleteURLs(ctx context.Context, userID string, urlsIDs []string) error {
	inputCh := make(chan string)
	defer close(inputCh)

	for _, v := range urlsIDs {
		inputCh <- v
	}

	modelToUpdateCh := fanIn(ctx, fanOut(ctx, inputCh, 10)...)
	return s.repository.UpdateURLsDeletedFlag(ctx, userID, modelToUpdateCh)
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

	// понадобится для ожидания всех горутин
	var wg sync.WaitGroup

	// перебираем все входящие каналы
	for _, ch := range resultChs {
		// в горутину передавать переменную цикла нельзя, поэтому делаем так
		chClosure := ch

		// инкрементируем счётчик горутин, которые нужно подождать
		wg.Add(1)

		go func() {
			// откладываем сообщение о том, что горутина завершилась
			defer wg.Done()

			// получаем данные из канала
			for data := range chClosure {
				select {
				// выходим из горутины, если канал закрылся
				case <-ctx.Done():
					return
				// если не закрылся, отправляем данные в конечный выходной канал
				case finalCh <- data:
				}
			}
		}()
	}

	go func() {
		// ждём завершения всех горутин
		wg.Wait()
		// когда все горутины завершились, закрываем результирующий канал
		close(finalCh)
	}()

	// возвращаем результирующий канал
	return finalCh
}
