package config

const (
	defaultAddress = ":8080"
	defaultBaseURL = "http://localhost:8080"
)

type Config struct {
	Address string // адрес запуска сервера
	BaseURL string // базовый адрес результирующего сокращённого URL
}

var cfg Config

func InitConfig() {
	cfg = Config{
		Address: defaultAddress,
		BaseURL: defaultBaseURL,
	}
}

func GetConfig() *Config {
	return &cfg
}
