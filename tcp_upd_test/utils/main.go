package utils

import "fmt"

func CreateServerAddress(addr string, port int) string {
	return fmt.Sprintf("%s:%d", addr, port)
}
