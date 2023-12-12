package config

import (
	"flag"
	"os"
)

const (
	defaultAddress = ":8080"
	defaultBaseURL = "http://localhost:8080"
)

type Config struct {
	// Адрес запуска сервера
	Address string
	// Базовый адрес результирующего сокращённого URL
	BaseURL string
}

var cfg Config

func InitConfig() {
	cfg.ParseFlags()
	cfg.ParseEnv()

	if cfg.Address == "" {
		cfg.Address = defaultAddress
	}
	if cfg.BaseURL == "" {
		cfg.Address = defaultBaseURL
	}
}

func SetConfig(c Config) {
	cfg = c
}

func GetConfig() *Config {
	return &cfg
}

func (cfg *Config) ParseFlags() {
	address := flag.String("a", defaultAddress, "server address; example: -a localhost:8080")
	baseURL := flag.String("b", defaultBaseURL, "short url base; example: -b https://yandex.ru")

	flag.Parse()

	cfg.Address = *address
	cfg.BaseURL = *baseURL
}

func (cfg *Config) ParseEnv() {
	if envAddress, ok := os.LookupEnv("SERVER_ADDRESS"); ok {
		cfg.Address = envAddress
	}

	if envBaseURL, ok := os.LookupEnv("BASE_URL"); ok {
		cfg.BaseURL = envBaseURL
	}
}
