package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"tcp_upd_test/tcp"
	customHttp "tcp_upd_test/tcp/http"
	"tcp_upd_test/udp"
)

func main() {

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if len(os.Args) < 3 {
		fmt.Println("Usage: tcp_upd_test <protocol type ('tcp', 'udp', 'http')> <port number ('8080', '54321')>")
		os.Exit(1)
	}
	protocolType := os.Args[1]
	port, portConvErr := strconv.Atoi(os.Args[2])
	if portConvErr != nil {
		fmt.Println("Port should be an integer")
		os.Exit(1)
	}

	switch protocolType {
	case "tcp":
		err := tcp.StartTCPServer(ctx, "", port)
		if err != nil {
			log.Fatal(err)
		}
	case "udp":
		err := udp.StartUDPServer(ctx, "", port)
		if err != nil {
			log.Fatal(err)
		}
	case "http":
		err := customHttp.StartHTTPServer(ctx, "", port)
		if err != nil {
			log.Fatal(err)
		}
	default:
		fmt.Println("Unknown protocol type:", protocolType)
		os.Exit(1)
	}

}
