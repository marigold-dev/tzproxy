package util

import (
	"os"
	"strconv"
)

func GetEnv(key string, defaultValue string) string {
	stringValue := os.Getenv(key)
	if stringValue == "" {
		return defaultValue
	}
	return stringValue
}

func GetEnvFloat(key string, defaultValue float64) float64 {
	stringValue := os.Getenv(key)
	if stringValue == "" {
		return defaultValue
	}
	value, err := strconv.ParseFloat(stringValue, 64)
	if err != nil {
		return defaultValue
	}
	return value
}

func GetEnvInt(key string, defaultValue int) int {
	stringValue := os.Getenv(key)
	if stringValue == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(stringValue)
	if err != nil {
		return defaultValue
	}
	return value
}
