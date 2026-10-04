package utils

import (
	"os"
	"strconv"
	"time"
	
	"github.com/joho/godotenv"
)

func LoadDotEnv(files ...string) error {
	if len(files) == 0 {
		files = []string{".env"}
	}
	for _, f := range files {
		if _, err := os.Stat(f); err != nil {
			continue
		}
		if err := godotenv.Load(f); err != nil {
			return err
		}
	}
	
	return nil
}

func GetEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func GetEnvInt(key string, fallback int) int {
	if n, err := strconv.Atoi(GetEnv(key, "")); err == nil {
		return n
	}
	return fallback
}

func GetEnvBool(key string, fallback bool) bool {
	if b, err := strconv.ParseBool(GetEnv(key, "")); err == nil {
		return b
	}
	return fallback
}

func GetEnvDuration(key string, fallback time.Duration) time.Duration {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback
	}
	
	if d, err := time.ParseDuration(v); err == nil {
		return d
	}
	if secs, err := strconv.Atoi(v); err == nil {
		return time.Duration(secs) * time.Second
	}
	return fallback
}
