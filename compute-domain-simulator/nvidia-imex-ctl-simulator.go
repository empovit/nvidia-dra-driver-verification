/*
 * Copyright (c) 2025 NVIDIA CORPORATION.  All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"
)

const (
	defaultTimeout = 5 * time.Second
)

type Config struct {
	ConfigPath string
	Query      bool
}

func main() {
	var config Config
	flag.StringVar(&config.ConfigPath, "c", "", "Path to IMEX daemon config file")
	flag.BoolVar(&config.Query, "q", false, "Query the status of the IMEX daemon")
	flag.Parse()

	if config.ConfigPath == "" {
		fmt.Fprintf(os.Stderr, "Error: config file path is required (-c flag)\n")
		os.Exit(1)
	}

	if config.Query {
		if err := queryStatus(&config); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Println("NVIDIA IMEX Control Simulator")
		fmt.Println("Available commands:")
		fmt.Println("  -q    Query daemon status")
		fmt.Println("  -c    Config file path (required)")
	}
}

func queryStatus(config *Config) error {
	// Parse the config file to get command service address
	cmdAddr, err := parseCommandAddress(config.ConfigPath)
	if err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	// Connect to the IMEX daemon command service
	conn, err := net.DialTimeout("tcp", cmdAddr, defaultTimeout)
	if err != nil {
		return fmt.Errorf("failed to connect to IMEX daemon at %s: %w", cmdAddr, err)
	}
	defer conn.Close()

	// Set read deadline
	conn.SetReadDeadline(time.Now().Add(defaultTimeout))

	// Read the status response
	reader := bufio.NewReader(conn)
	status, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return fmt.Errorf("failed to read status: %w", err)
	}

	status = strings.TrimSpace(status)

	// Print the status (daemon checks for "READY\n")
	fmt.Println(status)

	// Return error if not ready
	if status != "READY" {
		return fmt.Errorf("daemon is not ready: %s", status)
	}

	return nil
}

func parseCommandAddress(configPath string) (string, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return "", fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	var enabled bool
	var bindIP string
	var port int = 50005 // default

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "IMEX_CMD_ENABLED":
			enabled = (value == "1")
		case "IMEX_CMD_BIND_INTERFACE_IP":
			bindIP = value
		case "IMEX_CMD_PORT":
			fmt.Sscanf(value, "%d", &port)
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading config file: %w", err)
	}

	if !enabled {
		return "", fmt.Errorf("command service is not enabled in config")
	}

	if bindIP == "" {
		bindIP = "127.0.0.1"
	}

	return fmt.Sprintf("%s:%d", bindIP, port), nil
}

