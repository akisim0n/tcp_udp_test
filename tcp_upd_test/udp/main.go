package udp

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
)

const bufSize = 4096

func StartUDPServer(ctx context.Context, addr string, port int) error {
	udpConn, updErr := net.ListenPacket("udp", fmt.Sprintf("%s:%d", addr, port))
	if updErr != nil {
		log.Println("Error starting UDP Server:", updErr)
		return updErr
	}

	go func() {
		<-ctx.Done()
		udpConn.Close()
	}()

	connErr := handleUDPConnection(ctx, udpConn)
	if connErr != nil {
		log.Println("Error handling UDP connection:", connErr)
		return connErr
	}

	return nil
}

func handleUDPConnection(ctx context.Context, conn net.PacketConn) error {
	buf := make([]byte, bufSize)

	fmt.Println("Starting UDP Server")

	defer func() {
		fmt.Println("End UDP Server")
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			n, _, readErr := conn.ReadFrom(buf)
			if readErr != nil {
				log.Println(readErr)
			}
			_, err := os.Stdout.Write(buf[:n])
			if err != nil {
				log.Println(err)
			}
		}
	}
}
