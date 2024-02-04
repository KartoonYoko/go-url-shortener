package repository

import (
	"encoding/base64"
	"hash"
)

// var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// func randStringRunes(n int) string {
// 	b := make([]rune, n)
// 	for i := range b {
// 		b[i] = letterRunes[rand.Intn(len(letterRunes))]
// 	}
// 	return string(b)
// }

func GenerateURLUniqueHash(h hash.Hash, url string) (string, error) {
	h.Reset()
	_, err := h.Write([]byte(url))
	if err != nil {
		return "", err
	}
	hash := base64.URLEncoding.EncodeToString(h.Sum(nil)[:5])

	return hash, nil
}
