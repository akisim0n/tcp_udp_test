package tcp

import (
	"context"
	"fmt"
	"net"
	"os"
	"tcp_upd_test/utils"
)

const bufSize = 4096

func StartTCPServer(ctx context.Context, addr string, port int) error {
	tcpLis, tcpErr := net.Listen("tcp", utils.CreateServerAddress(addr, port))
	if tcpErr != nil {
		return tcpErr
	}

	go func() {
		<-ctx.Done()
		tcpLis.Close()
	}()

	for {
		conn, err := tcpLis.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			continue
		}
		go handleTCPConnection(ctx, conn)
	}
}

func handleTCPConnection(ctx context.Context, conn net.Conn) {
	fmt.Printf("Connection started: %s\n", conn.RemoteAddr())
	defer func(conn net.Conn) {
		fmt.Printf("Connection closed: %s\n", conn.RemoteAddr())
		conn.Close()
	}(conn)

	buf := make([]byte, bufSize)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			n, err := conn.Read(buf)
			if err != nil {
				return
			}

			_, err = os.Stdout.Write(buf[:n])
			if err != nil {
				return
			}
		}
	}
}
