package service

import (
	"errors"
	"math/rand"
	"time"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

var ErrNotFoundKey = errors.New("service: key not found")

// заглушка сервиса для создания коротких адресов
type Shortener struct {
	// хранилище адресов и их id'шников; ключ - id, значение - url
	Storage map[string]string
	r       *rand.Rand
}

func New() *Shortener {
	r := rand.New(rand.NewSource(time.Now().UnixMilli()))
	s := make(map[string]string)
	return &Shortener{
		Storage: s,
		r:       r,
	}
}

// сохранит url и вернёт его id'шник
func (s *Shortener) SaveURL(url string) string {
	hash := randStringRunes(5)
	s.Storage[hash] = url
	return hash
}

func (s *Shortener) GetURLByID(id string) (string, error) {
	res := s.Storage[id]

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
