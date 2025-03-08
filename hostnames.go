package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func loadHostnames(filename string) error {
	directory := filepath.Dir(filename)
	err := os.MkdirAll(directory, 0755)
	if err != nil {
		return fmt.Errorf("could not create directory %s: %v", directory, err)
	}

	fp, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0600)
	if err != nil {
		return fmt.Errorf("failed opening file with valid hostnames: %w", err)
	}

	defer fp.Close()

	// Add current list of hosts
	hosts, err := getHostnamesFromFile(fp)
	if err != nil {
		return err
	}

	validHostnames.Clear()
	for host, port := range *hosts {
		validHostnames.Store(host, port)
	}

	log.Printf("Loaded %d valid hostnames from file %s\n", len(*hosts), filename)
	return nil
}

func getPortFromHostname(hostname string) int {
	port, ok := validHostnames.Load(hostname)
	if !ok {
		return 0
	}
	return port.(int)
}

func getHostnamesFromFile(fp *os.File) (*map[string]int, error) {
	reader := bufio.NewReader(fp)
	hosts := make(map[string]int, 0)
	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				return &hosts, nil
			}
			return nil, err
		}

		trimmed := strings.TrimSpace(string(line))
		if len(trimmed) == 0 || trimmed[0] == '#' {
			continue
		}

		parts := strings.Split(trimmed, ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid format of proxy destination host. Line must have format with 'hostname:port'. Malformed line was: '%s'", trimmed)
		}

		port, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("invalid format of proxy destination host. Line must have format with 'hostname:port'. Malformed line was: '%s' and had error %w", trimmed, err)
		}

		if port <= 0 || port > 65535 {
			return nil, fmt.Errorf("invalid format of proxy destination host. Port must be in range 1-65535. Malformed line was: '%s'", trimmed)
		}

		hosts[parts[0]] = port
	}
}
