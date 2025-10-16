package config

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	ServVersion  string
	ServPort     string
	GracePeriod  time.Duration
	DatabasePath string
}

func LoadEnv() {
	envPath := ".env"

	absPath, err := filepath.Abs(envPath)
	if err != nil {
		log.Fatalf("Can not determined absolute path %s - %v", envPath, err)
	}

	_, statErr := os.Stat(envPath)
	if os.IsNotExist(statErr) {
		log.Fatalf(".env not found by path - %s", absPath)
	} else if statErr != nil {
		log.Fatalf("Error when trying to access .env: %v", statErr)
	} else {
		log.Printf(".env файл найден по пути: %s\n", absPath)
	}

	loadErr := godotenv.Load(envPath)
	if loadErr != nil {
		log.Fatalf("Failed to load .env - (%s): %v", absPath, loadErr)
	}

	log.Println(".env succesfully load.")

}

func GetConfig() *Config {

	graceStr := os.Getenv("GRACE_PERIOD")
	if graceStr == "" {
		graceStr = "5"
	}

	graceSec, err := strconv.Atoi(graceStr)
	if err != nil {
		log.Fatalf("Invalid GRACE_PERIOD: %v", err)
	}

	cfg := Config{
		ServVersion:  tryGetOrSetDefault("SERV_VERSION", "0.1.0"),
		ServPort:     tryGetOrSetDefault("SERV_PORT", ":8080"),
		GracePeriod:  time.Duration(graceSec) * time.Second,
		DatabasePath: tryGetOrSetDefault("DP_PATH", "db/rates.db"),
	}

	return &cfg
}

func tryGetOrSetDefault(key, def string) string {

	if value := os.Getenv(key); value != "" {
		return value
	}

	return def
}
