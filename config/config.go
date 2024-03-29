package config

import (
	"flag"
	"os"
)

type Config struct {
	// Адрес запуска сервера
	BootstrapNetAddress string
	// Базовый адрес результирующего сокращенного URL
	BaseURLAddress string
	// Полное имя файла, куда сохраняются сокращенные URL
	FileStoragePath string
	// строка подключения к БД
	DatabaseDsn string
}

func New() *Config {
	a := flag.String("a", ":8080", "Flag responsible for http server start")
	b := flag.String("b", "http://localhost:8080", "Flag responsible for base addres of shorted url")
	f := flag.String("f", "", "Path of short url's file")
	dbDsn := flag.String("d", "", "Database connection string")
	flag.Parse()

	c := &Config{
		BootstrapNetAddress: *a,
		BaseURLAddress:      *b,
		FileStoragePath:     *f,
		DatabaseDsn:         *dbDsn,
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

	return c
}
