package config

import (
	"io/fs"
	"log"
	"os"

	"github.com/joho/godotenv"
)

const (
	writeMode = fs.FileMode(0755)
)

// Загружает параметры из .env файла
func init() {
	if err := godotenv.Load(); err != nil {
		log.Print("Файл .env не был найден")
	}
}

// Получает переменную окружения по ключу
func getEnv(key string, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
