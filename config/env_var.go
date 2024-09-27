package config

import (
	"github.com/joho/godotenv"
	"os"
)

func getEnvVar(key string) string {
	err := godotenv.Load()
	if err != nil {
		//panic("load .env file error")
	}

	return os.Getenv(key)
}
