package config

import (
	"flag"
	"go.uber.org/zap/zapcore"
	"os"
)

const (
	defaultAddress         = ":8080"
	defaultBaseURL         = "http://localhost:8080"
	defaultLogLevel        = "info"
	defaultFileStoragePath = "/tmp/short-url-db.json"
)

type Config struct {
	// Адрес запуска сервера
	Address string
	// Базовый адрес результирующего сокращённого URL
	BaseURL string
	// Уровень логирования
	LogLevel string
	// Полное имя файла сохранения сокращенных URL
	FileStoragePath string
	// DSN подключения к бд
	DatabaseDSN string
}

func (cfg *Config) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddString("address", cfg.Address)
	encoder.AddString("base url", cfg.BaseURL)
	encoder.AddString("log level", cfg.LogLevel)
	encoder.AddString("storage file", cfg.FileStoragePath)
	encoder.AddString("database dsn", cfg.DatabaseDSN)

	return nil
}

func InitConfig() *Config {
	var cfg Config

	cfg.ParseFlags()
	cfg.ParseEnv()
	cfg.FillEmptyWithDefault()

	return &cfg
}

func (cfg *Config) ParseFlags() {
	address := flag.String("a", defaultAddress, "server address; example: -a localhost:8080")
	baseURL := flag.String("b", defaultBaseURL, "short url base; example: -b https://yandex.ru")
	logLevel := flag.String("l", defaultLogLevel, "log level; example: -l error")
	fileStoragePath := flag.String("f", defaultFileStoragePath, "file storage path; example: -f /home/pluhe7/file.json")
	databaseDSN := flag.String("d", "", "data source name for db; example: -d host=host port=port user=myuser password=xxxx dbname=mydb sslmode=disable")

	flag.Parse()

	cfg.Address = *address
	cfg.BaseURL = *baseURL
	cfg.LogLevel = *logLevel
	cfg.FileStoragePath = *fileStoragePath
	cfg.DatabaseDSN = *databaseDSN
}

func (cfg *Config) ParseEnv() {
	if envAddress, ok := os.LookupEnv("SERVER_ADDRESS"); ok {
		cfg.Address = envAddress
	}

	if envBaseURL, ok := os.LookupEnv("BASE_URL"); ok {
		cfg.BaseURL = envBaseURL
	}

	if envLogLevel, ok := os.LookupEnv("LOG_LEVEL"); ok {
		cfg.LogLevel = envLogLevel
	}

	if envFileStoragePath, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok {
		cfg.FileStoragePath = envFileStoragePath
	}

	if envDatabaseDSN, ok := os.LookupEnv("DATABASE_DSN"); ok {
		cfg.DatabaseDSN = envDatabaseDSN
	}
}

func (cfg *Config) FillEmptyWithDefault() {
	if cfg.Address == "" {
		cfg.Address = defaultAddress
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = defaultBaseURL
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = defaultLogLevel
	}
	if cfg.FileStoragePath == "" {
		cfg.FileStoragePath = defaultFileStoragePath
	}
}
