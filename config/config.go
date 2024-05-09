/*
Package config предоставляет конфигурацию приложения
*/
package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
)

// Config конфигурация приложения
type Config struct {
	// Адрес запуска сервера; флаг a
	BootstrapNetAddress string
	// Базовый адрес результирующего сокращенного URL; флаг b
	BaseURLAddress string
	// Полное имя файла, куда сохраняются сокращенные URL; флаг f
	FileStoragePath string
	// строка подключения к БД; флаг d
	DatabaseDsn string
	// Флаг активации HTTPS; флаг s
	EnableHTTPS bool
	// Путь к файлу конфигурации; флаг c
	ConfigFileName string
	// Разрешенные IP-адреса для некоторых ручек; флаг t
	TrustedSubnets string

	wasSetBootstrapNetAddress bool
	wasSetBaseURLAddress      bool
	wasSetFileStoragePath     bool
	wasSetDatabaseDsn         bool
	wasSetEnableHTTPS         bool
}

type configFileJSON struct {
	ServerAddress   *string `json:"server_address"`    // аналог переменной окружения SERVER_ADDRESS или флага -a
	BaseURL         *string `json:"base_url"`          // аналог переменной окружения BASE_URL или флага -b
	FileStoragePath *string `json:"file_storage_path"` // аналог переменной окружения FILE_STORAGE_PATH или флага -f
	DatabaseDSN     *string `json:"database_dsn"`      // аналог переменной окружения DATABASE_DSN или флага -d
	EnableHTTPS     *bool   `json:"enable_https"`      // аналог переменной окружения ENABLE_HTTPS или флага -s
	TrustedSubnets  *string `json:"trusted_subnet"`    // аналог переменной окружения TRUSTED_SUBNETS или флага -t
}

// New собирает конфигурацию из флагов командной строки, переменных среды
func New() (*Config, error) {
	c := new(Config)

	err := c.setFromFlags()
	if err != nil {
		return nil, err
	}

	err = c.setFromEnv()
	if err != nil {
		return nil, err
	}

	err = c.setFromConfigFile()
	if err != nil {
		return nil, err
	}

	return c, nil
}

// setFromEnv устанавливает данные из переменных окружения, если они не были заданы ранее
func (c *Config) setFromEnv() error {
	if !c.wasSetBootstrapNetAddress {
		envValue, ok := os.LookupEnv("SERVER_ADDRESS")
		c.wasSetBootstrapNetAddress = ok
		if ok {
			c.BootstrapNetAddress = envValue
		}
	}

	if !c.wasSetBaseURLAddress {
		envValue, ok := os.LookupEnv("BASE_URL")
		c.wasSetBaseURLAddress = ok
		if ok {
			c.BaseURLAddress = envValue
		}
	}

	if !c.wasSetFileStoragePath {
		envValue, ok := os.LookupEnv("FILE_STORAGE_PATH")
		c.wasSetFileStoragePath = ok
		if ok {
			c.FileStoragePath = envValue
		}
	}

	if !c.wasSetDatabaseDsn {
		envValue, ok := os.LookupEnv("DATABASE_DSN")
		c.wasSetDatabaseDsn = ok
		if ok {
			c.DatabaseDsn = envValue
		}
	}

	if !c.wasSetEnableHTTPS {
		envValue, ok := os.LookupEnv("ENABLE_HTTPS")
		c.wasSetEnableHTTPS = ok

		if ok {
			value, err := strconv.ParseBool(envValue)
			if err != nil {
				return err
			}
			c.EnableHTTPS = value
		}
	}

	return nil
}

// setFromFlags устанавливает данные из переданных флагов или данные по умолчанию
func (c *Config) setFromFlags() error {
	a := flag.String("a", ":8080", "Flag responsible for http server start")
	b := flag.String("b", "http://localhost:8080", "Flag responsible for base addres of shorted url")
	f := flag.String("f", "", "Path of short url's file")
	d := flag.String("d", "", "Database connection string")
	cf := flag.String("c", "", "Config file path")
	s := flag.Bool("s", false, "Enable TLS")
	flag.Parse()

	c.BootstrapNetAddress = *a
	c.BaseURLAddress = *b
	c.FileStoragePath = *f
	c.DatabaseDsn = *d
	c.EnableHTTPS = *s
	c.ConfigFileName = *cf

	c.wasSetBaseURLAddress = isFlagPassed("b")
	c.wasSetBootstrapNetAddress = isFlagPassed("a")
	c.wasSetDatabaseDsn = isFlagPassed("d")
	c.wasSetEnableHTTPS = isFlagPassed("s")
	c.wasSetFileStoragePath = isFlagPassed("f")

	return nil
}

// setFromConfigFile устанавливает только те данные из файла, которые ещё не заданы
func (c *Config) setFromConfigFile() error {
	if c.ConfigFileName == "" {
		return nil
	}
	f, err := os.Open(c.ConfigFileName)
	if err != nil {
		return fmt.Errorf("can not open config file: %w", err)
	}
	defer f.Close()

	var b []byte
	b, err = io.ReadAll(f)
	if err != nil {
		return fmt.Errorf("can not read config file: %w", err)
	}

	var j configFileJSON
	err = json.Unmarshal(b, &j)
	if err != nil {
		return fmt.Errorf("can not unmarshal config file: %w", err)
	}

	if !c.wasSetBaseURLAddress && j.BaseURL != nil {
		c.BaseURLAddress = *j.BaseURL
		c.wasSetBaseURLAddress = true
	}
	if !c.wasSetBootstrapNetAddress && j.ServerAddress != nil {
		c.BootstrapNetAddress = *j.ServerAddress
		c.wasSetBootstrapNetAddress = true
	}
	if !c.wasSetFileStoragePath && j.FileStoragePath != nil {
		c.FileStoragePath = *j.FileStoragePath
		c.wasSetFileStoragePath = true
	}
	if !c.wasSetDatabaseDsn && j.DatabaseDSN != nil {
		c.DatabaseDsn = *j.DatabaseDSN
		c.wasSetDatabaseDsn = true
	}
	if !c.wasSetEnableHTTPS && j.EnableHTTPS != nil {
		c.EnableHTTPS = *j.EnableHTTPS
		c.wasSetEnableHTTPS = true
	}
	return nil
}

// isFlagPassed определяет был ли передан флаг
func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}
