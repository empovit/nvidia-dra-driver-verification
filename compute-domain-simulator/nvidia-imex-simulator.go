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
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

// Version tracking - set at build time via ldflags
// SimulatorVersion matches TargetDriverVersion exactly
var (
	SimulatorVersion    = "v25.8.0-dev"
	TargetDriverVersion = "v25.8.0-dev"
)

const (
	defaultLogLevel   = 4
	defaultServerPort = 50000
	defaultCmdPort    = 50005
	readyState        = "READY"
	// Simulate initialization delay
	initDelay = 2 * time.Second
)

type Config struct {
	ConfigPath            string
	LogLevel              int
	LogFileName           string
	StatsFileName         string
	BindInterfaceIP       string
	ServerPort            int
	IMEXNodeConfigFile    string
	NetworkInterface      string
	IMEXWaitForQuorum     string
	IMEXCmdEnabled        bool
	IMEXCmdBindInterface  string
	IMEXCmdPort           int
	IMEXCmdUnixDomainPath string
	Daemonize             bool
}

type IMEXDaemon struct {
	config      *Config
	state       string
	stateLock   sync.RWMutex
	peers       []string
	listener    net.Listener
	cmdListener net.Listener
	shutdown    chan struct{}
}

func main() {
	configPath := flag.String("c", "", "Path to config file")
	flag.Parse()

	if *configPath == "" {
		log.Fatal("Config file path is required (-c flag)")
	}

	config, err := parseConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}

	daemon := NewIMEXDaemon(config)

	// Use plain context - daemon handles signals internally
	ctx := context.Background()

	if err := daemon.Run(ctx); err != nil {
		log.Fatalf("Daemon failed: %v", err)
	}
}

func NewIMEXDaemon(config *Config) *IMEXDaemon {
	return &IMEXDaemon{
		config:   config,
		state:    "INITIALIZING",
		shutdown: make(chan struct{}),
	}
}

func (d *IMEXDaemon) Run(ctx context.Context) error {
	// Ignore ctx passed from signal.NotifyContext in main, use our own signal handling
	d.log("Starting NVIDIA IMEX Daemon Simulator")
	d.log("Simulator version: %s (targeting driver %s)", SimulatorVersion, TargetDriverVersion)
	d.log("Config file: %s", d.config.ConfigPath)

	// Perform file system operations that the real driver would do
	// This tests security contexts, SCCs, and file permissions
	if err := d.initializeFileSystem(); err != nil {
		return fmt.Errorf("failed to initialize file system: %w", err)
	}

	// Load peers from node config file
	if err := d.loadPeers(); err != nil {
		return fmt.Errorf("failed to load peers: %w", err)
	}

	// Simulate initialization
	d.log("Initializing IMEX daemon...")
	time.Sleep(initDelay)

	// Start command/control service if enabled
	if d.config.IMEXCmdEnabled {
		if err := d.startCommandService(); err != nil {
			return fmt.Errorf("failed to start command service: %w", err)
		}
	}

	// Start main IMEX service
	if err := d.startMainService(); err != nil {
		return fmt.Errorf("failed to start main service: %w", err)
	}

	// Mark as ready
	d.setState(readyState)
	d.log("IMEX daemon is %s", readyState)
	d.log("Listening on %s:%d", d.config.BindInterfaceIP, d.config.ServerPort)

	// Write initial stats (simulating real daemon behavior)
	if err := d.writeStats(); err != nil {
		d.log("Warning: failed to write stats: %v", err)
	}

	// Start periodic stats writer
	statsTicker := time.NewTicker(30 * time.Second)
	defer statsTicker.Stop()

	// Wait for shutdown signal - separate channel to avoid conflicts with signal.NotifyContext
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT, syscall.SIGUSR1)

	for {
		select {
		case <-ctx.Done():
			d.log("Context canceled, shutting down...")
			return d.Shutdown()
		case <-statsTicker.C:
			// Periodic stats update (like real daemon)
			if err := d.writeStats(); err != nil {
				d.log("Warning: failed to write stats: %v", err)
			}
		case sig := <-sigCh:
			if sig == syscall.SIGUSR1 {
				d.log("Received SIGUSR1, reloading configuration...")
				// Reload peers from config file
				if err := d.loadPeers(); err != nil {
					d.log("Warning: failed to reload peers: %v", err)
				} else {
					d.log("Configuration reloaded successfully")
					// Update stats after reload
					if err := d.writeStats(); err != nil {
						d.log("Warning: failed to write stats: %v", err)
					}
				}
				// Continue running after reload
				continue
			} else {
				d.log("Received signal %s, shutting down...", sig)
				return d.Shutdown()
			}
		}
	}
}

func (d *IMEXDaemon) startMainService() error {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", d.config.BindInterfaceIP, d.config.ServerPort))
	if err != nil {
		return fmt.Errorf("failed to start main service: %w", err)
	}
	d.listener = listener

	go d.handleMainService()
	return nil
}

func (d *IMEXDaemon) startCommandService() error {
	addr := fmt.Sprintf("%s:%d", d.config.IMEXCmdBindInterface, d.config.IMEXCmdPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to start command service on %s: %w", addr, err)
	}
	d.cmdListener = listener

	d.log("Command service listening on %s", addr)
	go d.handleCommandService()
	return nil
}

func (d *IMEXDaemon) handleMainService() {
	for {
		conn, err := d.listener.Accept()
		if err != nil {
			select {
			case <-d.shutdown:
				return
			default:
				d.log("Error accepting connection: %v", err)
				continue
			}
		}

		go d.handlePeerConnection(conn)
	}
}

func (d *IMEXDaemon) handleCommandService() {
	for {
		conn, err := d.cmdListener.Accept()
		if err != nil {
			select {
			case <-d.shutdown:
				return
			default:
				d.log("Error accepting command connection: %v", err)
				continue
			}
		}

		go d.handleCommandConnection(conn)
	}
}

func (d *IMEXDaemon) handlePeerConnection(conn net.Conn) {
	defer conn.Close()
	d.log("New peer connection from %s", conn.RemoteAddr())
	// Keep connection alive until shutdown
	<-d.shutdown
}

func (d *IMEXDaemon) handleCommandConnection(conn net.Conn) {
	defer conn.Close()

	// Simple command protocol - just return status
	state := d.getState()
	fmt.Fprintf(conn, "%s\n", state)
}

func (d *IMEXDaemon) loadPeers() error {
	if d.config.IMEXNodeConfigFile == "" {
		d.log("No node config file specified")
		return nil
	}

	file, err := os.Open(d.config.IMEXNodeConfigFile)
	if err != nil {
		if os.IsNotExist(err) {
			d.log("Node config file does not exist yet: %s", d.config.IMEXNodeConfigFile)
			return nil
		}
		return err
	}
	defer file.Close()

	var peers []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			peers = append(peers, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	d.peers = peers
	d.log("Loaded %d peer(s) from %s", len(peers), d.config.IMEXNodeConfigFile)
	for i, peer := range peers {
		d.log("  Peer %d: %s", i+1, peer)
	}

	return nil
}

func (d *IMEXDaemon) setState(state string) {
	d.stateLock.Lock()
	defer d.stateLock.Unlock()
	d.state = state
}

func (d *IMEXDaemon) getState() string {
	d.stateLock.RLock()
	defer d.stateLock.RUnlock()
	return d.state
}

func (d *IMEXDaemon) Shutdown() error {
	d.log("Shutting down IMEX daemon...")
	close(d.shutdown)

	if d.listener != nil {
		d.listener.Close()
	}
	if d.cmdListener != nil {
		d.cmdListener.Close()
	}

	// Clean up PID file
	pidFile := "/var/run/nvidia-imex.pid"
	if err := os.Remove(pidFile); err != nil {
		if !os.IsNotExist(err) {
			d.log("Warning: failed to remove PID file: %v", err)
		}
	} else {
		d.log("Removed PID file: %s", pidFile)
	}

	// Write final stats
	if err := d.writeStats(); err != nil {
		d.log("Warning: failed to write final stats: %v", err)
	}

	d.log("IMEX daemon stopped")
	return nil
}

func (d *IMEXDaemon) log(format string, args ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "[IMEX] %s %s\n", timestamp, message)

	// Also write to log file if configured
	if d.config.LogFileName != "" {
		logLine := fmt.Sprintf("[IMEX] %s %s\n", timestamp, message)
		// Append to log file (ignore errors to not break on permission issues during logging)
		if f, err := os.OpenFile(d.config.LogFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
			f.WriteString(logLine)
			f.Close()
		}
	}
}

// initializeFileSystem performs all file system operations the real IMEX daemon would do
// This ensures proper security context testing (SCCs, permissions, etc.)
func (d *IMEXDaemon) initializeFileSystem() error {
	d.log("Initializing file system (testing security context)...")

	// Create /var/log directory if it doesn't exist
	if err := os.MkdirAll("/var/log", 0755); err != nil {
		d.log("Warning: cannot create /var/log: %v", err)
	}

	// Create /var/run directory if it doesn't exist
	if err := os.MkdirAll("/var/run", 0755); err != nil {
		d.log("Warning: cannot create /var/run: %v", err)
	}

	// Write PID file (standard daemon practice)
	pidFile := "/var/run/nvidia-imex.pid"
	pid := fmt.Sprintf("%d\n", os.Getpid())
	if err := os.WriteFile(pidFile, []byte(pid), 0644); err != nil {
		d.log("Warning: cannot write PID file %s: %v", pidFile, err)
	} else {
		d.log("Created PID file: %s", pidFile)
	}

	// Create log file if configured
	if d.config.LogFileName != "" {
		if err := d.initializeLogFile(d.config.LogFileName); err != nil {
			return fmt.Errorf("failed to initialize log file: %w", err)
		}
	}

	// Create stats file
	if d.config.StatsFileName != "" {
		if err := d.initializeLogFile(d.config.StatsFileName); err != nil {
			return fmt.Errorf("failed to initialize stats file: %w", err)
		}
	}

	// Create Unix domain socket if configured
	if d.config.IMEXCmdUnixDomainPath != "" {
		// Just test that we can create it (we'll remove it immediately)
		dir := filepath.Dir(d.config.IMEXCmdUnixDomainPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("cannot create socket directory %s: %w", dir, err)
		}
		d.log("Socket directory accessible: %s", dir)
	}

	// Test write to config directory (the real daemon might update config)
	configDir := filepath.Dir(d.config.ConfigPath)
	testFile := filepath.Join(configDir, ".imex-test-write")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return fmt.Errorf("cannot write to config directory %s: %w", configDir, err)
	}
	os.Remove(testFile) // Clean up test file
	d.log("Config directory is writable: %s", configDir)

	// Test write to node config directory
	if d.config.IMEXNodeConfigFile != "" {
		nodeDir := filepath.Dir(d.config.IMEXNodeConfigFile)
		if err := os.MkdirAll(nodeDir, 0755); err != nil {
			return fmt.Errorf("cannot create node config directory %s: %w", nodeDir, err)
		}
		testFile := filepath.Join(nodeDir, ".imex-test-write")
		if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
			return fmt.Errorf("cannot write to node config directory %s: %w", nodeDir, err)
		}
		os.Remove(testFile) // Clean up test file
		d.log("Node config directory is writable: %s", nodeDir)
	}

	d.log("File system initialization complete")
	return nil
}

// initializeLogFile creates a log file with proper permissions
func (d *IMEXDaemon) initializeLogFile(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("cannot create log directory %s: %w", dir, err)
	}

	// Create or truncate the file
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("cannot create log file %s: %w", path, err)
	}
	defer f.Close()

	// Write header
	header := fmt.Sprintf("# NVIDIA IMEX Simulator Log\n# Started: %s\n# PID: %d\n",
		time.Now().Format(time.RFC3339), os.Getpid())
	if _, err := f.WriteString(header); err != nil {
		return fmt.Errorf("cannot write to log file %s: %w", path, err)
	}

	d.log("Initialized log file: %s", path)
	return nil
}

// writeStats writes statistics to the stats file (simulating real daemon)
func (d *IMEXDaemon) writeStats() error {
	if d.config.StatsFileName == "" {
		return nil
	}

	stats := fmt.Sprintf(`# NVIDIA IMEX Statistics (Simulated)
# Timestamp: %s
# Status: %s
# Peers: %d
# Uptime: %s
# Port: %d
# Command Port: %d

[CONNECTIONS]
Total Peers: %d
Active Connections: 0
Failed Connections: 0

[MEMORY]
Allocated: 0 MB
In Use: 0 MB

[OPERATIONS]
Imports: 0
Exports: 0
Errors: 0
`,
		time.Now().Format(time.RFC3339),
		d.getState(),
		len(d.peers),
		time.Since(time.Now()).String(),
		d.config.ServerPort,
		d.config.IMEXCmdPort,
		len(d.peers),
	)

	f, err := os.OpenFile(d.config.StatsFileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("cannot open stats file: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(stats); err != nil {
		return fmt.Errorf("cannot write stats: %w", err)
	}

	return nil
}

func parseConfig(configPath string) (*Config, error) {
	config := &Config{
		ConfigPath:        configPath,
		LogLevel:          defaultLogLevel,
		ServerPort:        defaultServerPort,
		IMEXCmdPort:       defaultCmdPort,
		IMEXCmdEnabled:    true,
		IMEXWaitForQuorum: "RECOVERY",
		Daemonize:         false,
	}

	file, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

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
		case "LOG_LEVEL":
			fmt.Sscanf(value, "%d", &config.LogLevel)
		case "LOG_FILE_NAME":
			config.LogFileName = value
		case "STATS_FILE_NAME":
			config.StatsFileName = value
		case "BIND_INTERFACE_IP":
			config.BindInterfaceIP = value
		case "SERVER_PORT":
			fmt.Sscanf(value, "%d", &config.ServerPort)
		case "IMEX_NODE_CONFIG_FILE":
			config.IMEXNodeConfigFile = value
		case "NETWORK_INTERFACE":
			config.NetworkInterface = value
		case "IMEX_WAIT_FOR_QUORUM":
			config.IMEXWaitForQuorum = value
		case "IMEX_CMD_ENABLED":
			config.IMEXCmdEnabled = (value == "1")
		case "IMEX_CMD_BIND_INTERFACE_IP":
			config.IMEXCmdBindInterface = value
		case "IMEX_CMD_PORT":
			fmt.Sscanf(value, "%d", &config.IMEXCmdPort)
		case "IMEX_CMD_UNIX_DOMAIN_PATH":
			config.IMEXCmdUnixDomainPath = value
		case "DAEMONIZE":
			config.Daemonize = (value == "1")
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	// If BIND_INTERFACE_IP is empty, try to determine from node config
	if config.BindInterfaceIP == "" {
		if ip, err := getLocalIP(); err == nil {
			config.BindInterfaceIP = ip
		} else {
			config.BindInterfaceIP = "0.0.0.0"
		}
	}

	// Validate required fields
	if config.IMEXNodeConfigFile == "" {
		return nil, fmt.Errorf("IMEX_NODE_CONFIG_FILE is required")
	}

	// Ensure the directory for node config file exists
	dir := filepath.Dir(config.IMEXNodeConfigFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	return config, nil
}

func getLocalIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
}
