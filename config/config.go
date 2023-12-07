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
}

func New() *Config {
	a := flag.String("a", ":8080", "Flag responsible for http server start")
	b := flag.String("b", "http://localhost:8080", "Flag responsible for base addres of shorted url")
	flag.Parse()

	c := &Config{
		BootstrapNetAddress: *a,
		BaseURLAddress:      *b,
	}

	if envServerAddr := os.Getenv("SERVER_ADDRESS"); envServerAddr != "" {
		c.BootstrapNetAddress = envServerAddr
	}

	if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" {
		c.BaseURLAddress = envBaseURL
	}

	return c
}
