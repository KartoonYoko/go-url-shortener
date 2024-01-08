package usecase

type ShortenerRepo interface {
	SaveURL(url string) (string, error)
	GetURLByID(id string) (string, error)
}

type Shortener struct {
	repository ShortenerRepo
}

func New(repo ShortenerRepo) *Shortener {
	return &Shortener{
		repository: repo,
	}
}

// сохранит url и вернёт его id'шник
func (s *Shortener) SaveURL(url string) (string, error) {
	return s.repository.SaveURL(url)
}

func (s *Shortener) GetURLByID(id string) (string, error) {
	return s.repository.GetURLByID(id)
}
