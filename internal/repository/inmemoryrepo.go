package repository

import (
	"math/rand"
	"time"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// хранилище коротки адресов в памяти
type InMemoryRepo struct {
	// хранилище адресов и их id'шников; ключ - id, значение - url
	storage map[string]string
	r       *rand.Rand
}

func NewInMemoryRepo() *InMemoryRepo {
	r := rand.New(rand.NewSource(time.Now().UnixMilli()))
	s := make(map[string]string)
	return &InMemoryRepo{
		storage: s,
		r:       r,
	}
}

// сохранит url и вернёт его id'шник
func (s *InMemoryRepo) SaveURL(url string) string {
	hash := randStringRunes(5)
	s.storage[hash] = url
	return hash
}

func (s *InMemoryRepo) GetURLByID(id string) (string, error) {
	res := s.storage[id]

	if res == "" {
		return res, ErrNotFoundKey
	}

	return res, nil
}

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}