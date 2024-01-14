package shortener

type ShortenerRepo interface {
	SaveURL(url string) (string, error)
	GetURLByID(id string) (string, error)
}

type shortenerUsecase struct {
	repository ShortenerRepo
}

func New(repo ShortenerRepo) *shortenerUsecase {
	return &shortenerUsecase{
		repository: repo,
	}
}

// сохранит url и вернёт его id'шник
func (s *shortenerUsecase) SaveURL(url string) (string, error) {
	return s.repository.SaveURL(url)
}

func (s *shortenerUsecase) GetURLByID(id string) (string, error) {
	return s.repository.GetURLByID(id)
}
