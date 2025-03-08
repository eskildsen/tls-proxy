package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"log"
	"strings"
	"sync"
)

func getArgs() (string, *tls.Config, string, *string) {
	listenPort := flag.Int("p", 1337, "port to listen on")
	listenHost := flag.String("host", "0.0.0.0", "host to listen on")
	certificate := flag.String("cert", "certs/certificate.pem", "the certificate to use for incoming TLS")
	certificateKey := flag.String("key", "certs/priv.key", "the corresponding certificate key")
	destFile := flag.String("dest", "targets.txt", "a file containing hostnames and corresponding local ports for end destination")
	metricsFile := flag.String("m", "", "filename for storing traffic metrics")

	flag.Parse()

	if destFile == nil {
		panic(fmt.Errorf("parameter 'dest' cannot be empty"))
	}

	tlsConf, err := loadCertificate(*certificate, *certificateKey)
	if err != nil {
		panic(fmt.Errorf("failed to load certificate: %w", err))
	}

	if *metricsFile == "" {
		metricsFile = nil
	}

	listenAddress := fmt.Sprintf("%s:%d", *listenHost, *listenPort)
	return listenAddress, tlsConf, *destFile, metricsFile
}

func loadCertificate(certFile, keyfile string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyfile)
	if err != nil {
		return nil, err
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
		MaxVersion:   tls.VersionTLS12,
	}
	return config, nil
}

func getHostNames(config *tls.Config) []string {
	raw := config.Certificates[0].Certificate[0] // leaf
	cert, err := x509.ParseCertificate(raw)
	if err != nil {
		log.Printf("Failed parsing certificate: %v", err)
		return []string{}
	}

	return cert.DNSNames
}

func listen(listenAddr string, conf *tls.Config) error {
	ln, err := tls.Listen("tcp", listenAddr, conf)
	if err != nil {
		return fmt.Errorf("failed to start proxy listener: %w", err)
	}
	defer ln.Close()

	hostnames := getHostNames(conf)
	log.Printf("Listening on %s with hostname(s): %s\n", listenAddr, strings.Join(hostnames, ", "))

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(fmt.Errorf("failed accepting client: %w", err))
			continue
		}

		go handleTLSHandshake(conn)
	}
}

var validHostnames sync.Map

func main() {
	listenAddress, tlsConf, hostsFile, metricsFile := getArgs()

	err := loadHostnames(hostsFile)
	if err != nil {
		panic(err)
	}

	go setupHostnameWatcher(hostsFile)

	if metricsFile != nil {
		if err = openMetrics(*metricsFile); err != nil {
			log.Printf("Failed to init metrics: %v", err)
		} else {
			defer func() {
				_ = closeMetrics()
			}()
		}
	} else {
		log.Printf("No metrics file provided. Metric logging disabled")
	}

	err = listen(listenAddress, tlsConf)
	if err != nil {
		panic(err)
	}
}
