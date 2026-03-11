package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"tcp_upd_test/tcp"
	"tcp_upd_test/udp"
)

func main() {

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	switch len(os.Args) {
	case 1:
		fmt.Println("Usage: tcp_upd_test <protocol type ('tcp', 'udp')> <port number ('8080', '54321')>")
		os.Exit(1)
	case 2:
		fmt.Println("Usage: tcp_upd_test <protocol type ('tcp', 'udp')> <port number ('8080', '54321')>")
		os.Exit(1)
	}
	if len(os.Args) < 3 {

	}
	protocolType := os.Args[1]

	switch protocolType {
	case "tcp":
		err := tcp.StartTCPServer(ctx, "", 8080)
		if err != nil {
			log.Fatal(err)
		}
	case "udp":
		err := udp.StartUDPServer(ctx, "", 8081)
		if err != nil {
			log.Fatal(err)
		}
	default:
		fmt.Println("Unknown protocol type:", protocolType)
		os.Exit(1)
	}

}
