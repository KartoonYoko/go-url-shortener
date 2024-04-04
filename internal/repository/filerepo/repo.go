package filerepo

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"os"
	"strconv"

	model "github.com/KartoonYoko/go-url-shortener/internal/model/shortener"
	inmr "github.com/KartoonYoko/go-url-shortener/internal/repository/inmemoryrepo"
)

// строка записи в файле
type recordShorURL struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	// UserID      string `json:"user_id"`
}

type fileRepo struct {
	// хранилище адресов и их id'шников; ключ - id, значение - данные
	repo         inmr.InMemoryRepo
	lineLastUUID int
	filename     string
	file         *os.File
}

// Конструктор для хранилища-файла
func NewFileRepo(fileName string) (*fileRepo, error) {
	repo := &fileRepo{
		repo:         *inmr.NewInMemoryRepo(),
		lineLastUUID: 0,
		filename:     fileName,
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
func (s *fileRepo) SaveURL(ctx context.Context, url string, userID string) (string, error) {
	hash, err := s.repo.SaveURL(ctx, url, userID)
	if err != nil {
		return "", err
	}
	record := recordShorURL{
		UUID:        strconv.FormatInt(int64(s.lineLastUUID+1), 10),
		ShortURL:    hash,
		OriginalURL: url,
		// UserID:      userID,
	}
	err = s.saveToFile(record)
	if err != nil {
		return "", err
	}

	s.lineLastUUID++
	return hash, nil
}

// GetURLByID вернёт URL по его ID
func (s *fileRepo) GetURLByID(ctx context.Context, id string) (string, error) {
	return s.repo.GetURLByID(ctx, id)
}

// GetUserURLs вернёт все URL'ы, которые пользователь создавал когда-либо
func (s *fileRepo) GetUserURLs(ctx context.Context, userID string) ([]model.GetUserURLsItemResponse, error) {
	return s.repo.GetUserURLs(ctx, userID)
}

// Close закрывает файл
func (s *fileRepo) Close() error {
	s.repo.Close()
	return s.file.Close()
}

// Ping релизует интерфейс Pinger
func (s *fileRepo) Ping(ctx context.Context) error {
	return s.repo.Ping(ctx)
}

// SaveURLsBatch сохраняет множество URL'ов
func (s *fileRepo) SaveURLsBatch(
	ctx context.Context,
	request []model.CreateShortenURLBatchItemRequest,
	userID string) ([]model.CreateShortenURLBatchItemResponse, error) {
	response := make([]model.CreateShortenURLBatchItemResponse, len(request))
	for _, v := range request {
		hash, err := s.SaveURL(ctx, v.OriginalURL, userID)
		if err != nil {
			return nil, err
		}

		response = append(response, model.CreateShortenURLBatchItemResponse{
			CorrelationID: v.CorrelationID,
			ShortURL:      hash,
		})
	}

	return response, nil
}

// GetNewUserID сгенерировать новый ID для пользователя
func (s *fileRepo) GetNewUserID(ctx context.Context) (string, error) {
	return s.repo.GetNewUserID(ctx)
}

// UpdateURLsDeletedFlag пометит указанные URL'ы удалёнными
func (s *fileRepo) UpdateURLsDeletedFlag(ctx context.Context, userID string, modelsCh <-chan model.UpdateURLDeletedFlag) error {
	return s.repo.UpdateURLsDeletedFlag(ctx, userID, modelsCh)
}

func (s *fileRepo) loadAllData() error {
	if s.filename == "" {
		return nil
	}

	file, err := os.OpenFile(s.filename, os.O_RDONLY|os.O_CREATE, 0744)
	if err != nil {
		return err
	}

	ctx := context.TODO()
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

		_, err = s.repo.SaveURL(ctx, record.OriginalURL, "")
		if err != nil {
			return err
		}
		parsed, err := strconv.ParseInt(record.UUID, 10, 32)
		if err != nil {
			return err
		}
		if s.lineLastUUID < int(parsed) {
			s.lineLastUUID = int(parsed)
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
