package utils

import "os"

func GetEnv(key, default_ string) string {
	value := os.Getenv(key)
	if value == "" {
		return default_
	}
	return value
}
