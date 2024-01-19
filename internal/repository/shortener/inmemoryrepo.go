package shortener

import (
	"context"
	"math/rand"
	"time"
)

// хранилище коротки адресов в памяти
type inMemoryRepo struct {
	// хранилище адресов и их id'шников; ключ - id, значение - url
	storage map[string]string
	r       *rand.Rand
}

func NewInMemoryRepo() *inMemoryRepo {
	r := rand.New(rand.NewSource(time.Now().UnixMilli()))
	s := make(map[string]string)
	return &inMemoryRepo{
		storage: s,
		r:       r,
	}
}

// сохранит url и вернёт его id'шник
func (s *inMemoryRepo) SaveURL(url string) (string, error) {
	hash := randStringRunes(5)
	s.storage[hash] = url
	return hash, nil
}

func (s *inMemoryRepo) GetURLByID(id string) (string, error) {
	res := s.storage[id]

	if res == "" {
		return res, ErrNotFoundKey
	}

	return res, nil
}

func (s *inMemoryRepo) Ping(ctx context.Context) error {
	return nil
}

func (s *inMemoryRepo) Close() error {
	return nil
}
