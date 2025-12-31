package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	socketDir         = ".cdiff"
	sidebarSocketName = "cdiff.sock"
	groupingTimeout   = 2 * time.Second
)

// IPCCommand represents a command sent via IPC
type IPCCommand struct {
	Cmd   string    `json:"cmd"`   // "add-file"
	Entry FileEntry `json:"entry"` // the file to add
}

// IPCServer manages the Unix socket server for receiving commands
type IPCServer struct {
	listener     net.Listener
	socketPath   string
	app          *App
	mu           sync.Mutex
	closed       bool
	timeoutTimer *time.Timer
}

// getSocketDir returns the socket directory path
func getSocketDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = os.TempDir()
	}
	return filepath.Join(homeDir, socketDir)
}

// getSidebarSocketPath returns the path for the sidebar mode socket
func getSidebarSocketPath() string {
	return filepath.Join(getSocketDir(), sidebarSocketName)
}

// ensureSocketDir creates the socket directory if it doesn't exist
func ensureSocketDir() error {
	dir := getSocketDir()
	return os.MkdirAll(dir, 0700)
}

// TrySendToSidebarInstance tries to send a file to an existing sidebar instance
// Returns true if successful (caller should exit), false if no instance running
func TrySendToSidebarInstance(entry FileEntry) bool {
	socketPath := getSidebarSocketPath()
	conn, err := net.DialTimeout("unix", socketPath, 500*time.Millisecond)
	if err != nil {
		// Connection failed - socket might be stale, clean it up
		os.Remove(socketPath)
		return false
	}
	defer conn.Close()

	cmd := IPCCommand{
		Cmd:   "add-file",
		Entry: entry,
	}

	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(cmd); err != nil {
		return false
	}

	return true
}

// NewIPCServer creates a new IPC server
func NewIPCServer(app *App, socketPath string) (*IPCServer, error) {
	if err := ensureSocketDir(); err != nil {
		return nil, fmt.Errorf("failed to create socket directory: %w", err)
	}

	// Remove existing socket file if it exists
	os.Remove(socketPath)

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create socket: %w", err)
	}

	server := &IPCServer{
		listener:   listener,
		socketPath: socketPath,
		app:        app,
	}

	// Start timeout timer
	server.resetTimeout()

	return server, nil
}

// resetTimeout resets the grouping timeout timer
func (s *IPCServer) resetTimeout() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.timeoutTimer != nil {
		s.timeoutTimer.Stop()
	}

	s.timeoutTimer = time.AfterFunc(groupingTimeout, func() {
		s.Close()
	})
}

// Start begins accepting connections
func (s *IPCServer) Start() {
	go func() {
		for {
			conn, err := s.listener.Accept()
			if err != nil {
				s.mu.Lock()
				closed := s.closed
				s.mu.Unlock()
				if closed {
					return
				}
				continue
			}

			go s.handleConnection(conn)
		}
	}()
}

// handleConnection processes a single IPC connection
func (s *IPCServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	// Reset timeout on each new file
	s.resetTimeout()

	decoder := json.NewDecoder(conn)
	var cmd IPCCommand
	if err := decoder.Decode(&cmd); err != nil {
		return
	}

	if cmd.Cmd == "add-file" {
		s.app.AddFile(cmd.Entry)
	}
}

// Close shuts down the IPC server and removes the socket file
func (s *IPCServer) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return
	}
	s.closed = true

	if s.timeoutTimer != nil {
		s.timeoutTimer.Stop()
	}

	if s.listener != nil {
		s.listener.Close()
	}

	// Clean up socket file
	os.Remove(s.socketPath)
}

// StartSidebarServer starts an IPC server for sidebar mode with timeout
func StartSidebarServer(app *App) (*IPCServer, error) {
	server, err := NewIPCServer(app, getSidebarSocketPath())
	if err != nil {
		return nil, err
	}
	server.Start()
	return server, nil
}
