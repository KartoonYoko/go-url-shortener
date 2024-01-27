package shortener

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/KartoonYoko/go-url-shortener/internal/model"
)

// строка записи в файле
type recordShorURL struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type fileRepo struct {
	// хранилище адресов и их id'шников; ключ - id, значение - url
	storage  map[string]string
	r        *rand.Rand
	lastUUID int
	filename string
	file     *os.File
}

func NewFileRepo(fileName string) (*fileRepo, error) {
	r := rand.New(rand.NewSource(time.Now().UnixMilli()))
	s := make(map[string]string)

	repo := &fileRepo{
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
func (s *fileRepo) SaveURL(ctx context.Context, url string) (string, error) {
	hash := randStringRunes(5)
	record := recordShorURL{
		UUID:        strconv.FormatInt(int64(s.lastUUID+1), 10),
		ShortURL:    hash,
		OriginalURL: url,
	}
	err := s.saveToFile(record)
	if err != nil {
		return "", err
	}

	s.lastUUID++
	s.storage[hash] = url
	return hash, nil
}

func (s *fileRepo) GetURLByID(ctx context.Context, id string) (string, error) {
	res := s.storage[id]

	if res == "" {
		return res, ErrNotFoundKey
	}

	return res, nil
}

func (s *fileRepo) Close() error {
	return s.file.Close()
}

func (s *fileRepo) Ping(ctx context.Context) error {
	return nil
}

func (s *fileRepo) SaveURLsBatch(ctx context.Context,
	request []model.CreateShortenURLBatchItemRequest) ([]model.CreateShortenURLBatchItemResponse, error) {
	response := make([]model.CreateShortenURLBatchItemResponse, len(request))
	for _, v := range request {
		hash, err := s.SaveURL(ctx, v.OriginalURL)
		if err != nil {
			return nil, err
		}

		response = append(response, model.CreateShortenURLBatchItemResponse{
			CorrelationID: v.CorrelationID,
			ShortURL: hash,
		})
	}

	return response, nil
}

func (s *fileRepo) loadAllData() error {
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
		parsed, err := strconv.ParseInt(record.UUID, 10, 32)
		if err != nil {
			return err
		}
		if s.lastUUID < int(parsed) {
			s.lastUUID = int(parsed)
		}
	}

	return nil
}

func (s *fileRepo) saveToFile(r recordShorURL) error {
	data, err := json.Marshal(r)
	if err != nil {
		return err
	}

	// добавляем перенос строки
	data = append(data, '\n')

	_, err = s.file.Write(data)
	return err
}
