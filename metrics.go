package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"
)

type Metrics struct {
	ClientAddress string
	TargetAddress string
	Hostname      string
	Sent          int64
	Received      int64
	Connected     time.Time
	Disconnected  time.Time
}

var metricsFp *os.File

func openMetrics(filename string) error {
	// In the future me might ship these off to a remote host instead
	fp, err := os.OpenFile(filename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0600)

	if err != nil {
		return err
	}

	metricsFp = fp
	return nil
}

func addMetrics(m *Metrics) error {
	if metricsFp == nil {
		return errors.New("metrics file not opened")
	}

	serialized, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("failed writing metrics: %w", err)
	}

	_, err = metricsFp.Write(append(serialized, '\n'))
	if err != nil {
		return fmt.Errorf("failed to write metrics: %w", err)
	}

	return nil
}

func closeMetrics() error {
	if metricsFp == nil {
		return nil
	}

	err := metricsFp.Close()
	if err != nil {
		return fmt.Errorf("failed to close metrics: %w", err)
	}

	metricsFp = nil
	return nil
}
