package env

import "os"

func Get(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

func Required(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic(key + " is required")
	}

	return value
}
