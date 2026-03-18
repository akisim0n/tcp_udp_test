package utils

import (
	"fmt"
	"os"
)

func CreateServerAddress(addr string, port int) string {
	return fmt.Sprintf("%s:%d", addr, port)
}

func GetEnvParam(key string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return ""
}
