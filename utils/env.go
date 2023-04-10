package util

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"

	"golang.org/x/exp/constraints"
)

func GetEnv(key string, defaultValue string) string {
	stringValue := os.Getenv(key)
	if stringValue == "" {
		return defaultValue
	}
	return stringValue
}

func GetEnvBool(key string, defaultValue bool) bool {
	stringValue := os.Getenv(key)
	if stringValue == "" {
		return defaultValue
	}

	return strings.Contains(strings.ToLower(stringValue), "true")
}

func GetEnvFloat[T constraints.Float](key string, defaultValue T) T {
	stringValue := os.Getenv(key)
	if stringValue == "" {
		return defaultValue
	}
	value, err := strconv.ParseFloat(stringValue, 64)
	if err != nil {
		return defaultValue
	}
	return T(value)
}

func GetEnvInt[T constraints.Integer](key string, defaultValue T) T {
	stringValue := os.Getenv(key)
	if stringValue == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(stringValue)
	if err != nil {
		return defaultValue
	}
	return T(value)
}

func GetEnvSlice[T comparable](key string, defaultValue []T) []T {
	stringValue := os.Getenv(key)
	if stringValue == "" {
		return defaultValue
	}
	value := []T{}
	err := json.Unmarshal([]byte(stringValue), &value)
	if err != nil {
		return defaultValue
	}
	return value
}

func GetEnvSet[T comparable](key string, defaultValue map[T]bool) map[T]bool {
	stringValue := os.Getenv(key)
	if stringValue == "" {
		return defaultValue
	}
	value := []T{}
	err := json.Unmarshal([]byte(stringValue), &value)
	if err != nil {
		return defaultValue
	}
	set := map[T]bool{}
	for _, v := range value {
		set[v] = true
	}
	return set
}
