/*
Package config предоставляет конфигурацию приложения
*/
package config

import (
	"flag"
	"os"
	"strconv"
)

// Config конфигурация приложения
type Config struct {
	// Адрес запуска сервера
	BootstrapNetAddress string
	// Базовый адрес результирующего сокращенного URL
	BaseURLAddress string
	// Полное имя файла, куда сохраняются сокращенные URL
	FileStoragePath string
	// строка подключения к БД
	DatabaseDsn string
	EnableHTTPS bool
}

// New собирает конфигурацию из флагов командной строки, переменных среды
func New() (*Config, error) {
	a := flag.String("a", ":8080", "Flag responsible for http server start")
	b := flag.String("b", "http://localhost:8080", "Flag responsible for base addres of shorted url")
	f := flag.String("f", "", "Path of short url's file")
	d := flag.String("d", "", "Database connection string")
	s := flag.Bool("s", false, "Enable HTTPS ")
	flag.Parse()

	c := &Config{
		BootstrapNetAddress: *a,
		BaseURLAddress:      *b,
		FileStoragePath:     *f,
		DatabaseDsn:         *d,
		EnableHTTPS:         *s,
	}

	if envServerAddr := os.Getenv("SERVER_ADDRESS"); envServerAddr != "" {
		c.BootstrapNetAddress = envServerAddr
	}

	if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" {
		c.BaseURLAddress = envBaseURL
	}

	if envFileStoragePath := os.Getenv("FILE_STORAGE_PATH"); envFileStoragePath != "" {
		c.FileStoragePath = envFileStoragePath
	}

	if envDatabaseDsn := os.Getenv("DATABASE_DSN"); envDatabaseDsn != "" {
		c.DatabaseDsn = envDatabaseDsn
	}

	if envEnableHTTPS := os.Getenv("ENABLE_HTTPS"); envEnableHTTPS != "" {
		value, err := strconv.ParseBool(envEnableHTTPS)
		if err != nil {
			return nil, err
		}
		c.EnableHTTPS = value
	}

	return c, nil
}
