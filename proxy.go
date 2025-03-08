package main

import (
	"context"
	"crypto/tls"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

func handleTLSHandshake(conn net.Conn) {
	defer conn.Close()

	log.Printf("Got a new connection from: %s\n", conn.RemoteAddr())
	tlscon, ok := conn.(*tls.Conn)
	if !ok {
		log.Print("Failed casting connection to TLS - no idea why. Aborting connection")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := tlscon.HandshakeContext(ctx)

	if err != nil {
		log.Printf("Handshake failed, aborting connection: %v\n", err)
		return
	}

	state := tlscon.ConnectionState()
	log.Printf("Handshake completed, target server: %s\n", state.ServerName)

	port := getPortFromHostname(strings.ToLower(state.ServerName))
	if port == 0 {
		log.Printf("Unknown hostname '%s', aborting connection\n", state.ServerName)
		_, _ = tlscon.Write([]byte("Unknown hostname. Please reconnect using an known hostname. Your instance might have expired.\n"))
		return
	}

	handleConnection(tlscon, state.ServerName, port)
}

func handleConnection(conn net.Conn, hostname string, destPort int) {
	dest := net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: destPort}
	tcpDestConn, err := net.DialTCP("tcp", nil, &dest)

	if err != nil {
		log.Printf("Failed to connect to destionation server on port %d: %v\n", destPort, err)
		return
	}

	var bytesSent int64
	var bytesRecv int64
	connected := time.Now()

	defer func() {
		_ = addMetrics(&Metrics{
			ClientAddress: conn.RemoteAddr().String(),
			TargetAddress: tcpDestConn.LocalAddr().String(),
			Hostname:      hostname,
			Sent:          bytesSent,
			Received:      bytesRecv,
			Connected:     connected,
			Disconnected:  time.Now(),
		})
	}()

	// Now actually proxy traffic
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		bytesSent, err = io.Copy(conn, tcpDestConn)

		if err != nil {
			log.Printf("Error from DEST to CLIENT: %v\n", err)
		}

		log.Printf("DEST closed connection")
		// Signal peer that no more data is coming from DEST.
		conn.Close()
	}()
	go func() {
		defer wg.Done()
		bytesRecv, err = io.Copy(tcpDestConn, conn)

		if err != nil {
			log.Printf("Error from CLIENT to DEST: %v\n", err)
		}

		log.Printf("CLIENT closed connection")

		// Signal peer that no more data is coming.
		tcpDestConn.CloseWrite()
	}()

	wg.Wait()

	log.Printf("Connection terminating")
}
