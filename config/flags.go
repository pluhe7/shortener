package config

import (
	"flag"
)

func (cfg *Config) ParseFlags() {
	address := flag.String("a", defaultAddress, "server address; example: -a localhost:8080")
	baseURL := flag.String("b", defaultBaseURL, "short url base; example: -b https://yandex.ru")

	flag.Parse()

	cfg.Address = *address
	cfg.BaseURL = *baseURL
}
