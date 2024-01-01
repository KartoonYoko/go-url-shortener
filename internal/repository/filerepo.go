package repository

import (
	"bufio"
	"encoding/json"
	"io"
	"math/rand"
	"os"
	"strconv"
	"time"
)

// строка записи в файле
type recordShorURL struct {
	Uuid        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type FileRepo struct {
	// хранилище адресов и их id'шников; ключ - id, значение - url
	storage  map[string]string
	r        *rand.Rand
	lastUUID int
	filename string
	file     *os.File
}

func NewFileRepo(fileName string) (*FileRepo, error) {
	r := rand.New(rand.NewSource(time.Now().UnixMilli()))
	s := make(map[string]string)

	repo := &FileRepo{
		storage:  s,
		r:        r,
		lastUUID: 0,
		filename: fileName,
	}

	err := repo.loadAllData()
	if err != nil {
		return nil, err
	}

	repo.file, err = os.OpenFile(repo.filename, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	return repo, nil
}

// сохранит url и вернёт его id'шник
func (s *FileRepo) SaveURL(url string) (string, error) {
	hash := randStringRunes(5)
	record := recordShorURL{
		Uuid: strconv.FormatInt(int64(s.lastUUID+1), 10),
		ShortURL: hash,
		OriginalURL: url,
	}
	err := s.saveToFile(record)
	if err != nil {
		return "", err
	}

	s.storage[hash] = url
	return hash, nil
}

func (s *FileRepo) GetURLByID(id string) (string, error) {
	res := s.storage[id]

	if res == "" {
		return res, ErrNotFoundKey
	}

	return res, nil
}

func (s *FileRepo) loadAllData() error {
	if s.filename == "" {
		return nil
	}

	file, err := os.OpenFile(s.filename, os.O_RDONLY|os.O_CREATE, 0744)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(file)
	for {
		if !scanner.Scan() {
			return scanner.Err()
		}
		record := &recordShorURL{}
		data := scanner.Bytes()
		err := json.Unmarshal(data, record)

		if err != nil {
			if err == io.EOF {
				break
			}

			return err
		}

		s.storage[record.ShortURL] = record.OriginalURL
		parsed, err := strconv.ParseInt(record.Uuid, 10, 32)
		if err != nil {
			return err
		}
		if s.lastUUID < int(parsed) {
			s.lastUUID = int(parsed)
		}
	}

	return nil
}

func (s *FileRepo) saveToFile(r recordShorURL) error {
	data, err := json.Marshal(r)
	if err != nil {
		return err
	}

	_, err = s.file.Write(data)
	return err
}
