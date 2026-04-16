package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
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

func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
