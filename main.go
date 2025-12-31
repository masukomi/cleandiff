package main

import (
	"context"
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
)

//go:embed frontend/*
var assets embed.FS

const Version = "1.0.0"

var (
	internalGUI bool   // Hidden flag: run as GUI subprocess
	htmlContent string // Hidden flag: pre-generated HTML content path
)

func main() {
	// Check for internal GUI mode (called by parent process)
	if len(os.Args) >= 2 && os.Args[1] == "--internal-gui" {
		internalGUI = true
		if len(os.Args) >= 4 && os.Args[2] == "--html" {
			htmlContent = os.Args[3]
		}
		if len(os.Args) >= 5 {
			// Get display name from args
			runGUI(htmlContent, os.Args[4])
		} else {
			runGUI(htmlContent, "diff")
		}
		return
	}

	// Normal CLI invocation
	args := os.Args[1:]

	// Check for version flag
	if len(args) == 1 && (args[0] == "-v" || args[0] == "--version") {
		fmt.Printf("cdiff %s\n", Version)
		os.Exit(0)
	}

	// Check for debug flag
	debugMode := false
	filteredArgs := make([]string, 0, len(args))
	for _, arg := range args {
		if arg == "--debug" {
			debugMode = true
		} else {
			filteredArgs = append(filteredArgs, arg)
		}
	}
	args = filteredArgs

	// Validate arguments
	if len(args) < 2 || len(args) > 3 {
		printUsage()
		// Exit 0 for no args (help), exit 1 for wrong number of args
		if len(args) == 0 {
			os.Exit(0)
		}
		os.Exit(1)
	}

	file1Path := args[0]
	file2Path := args[1]
	var fileName string

	if len(args) == 3 {
		fileName = args[2]
	} else {
		// Derive filename from second file path
		// Strip first 7 characters from basename (git difftool temp file prefix)
		base := filepath.Base(file2Path)
		if len(base) > 7 {
			fileName = base[7:]
		} else {
			fileName = base
		}
	}

	// Read both files
	file1Content, err := os.ReadFile(file1Path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", file1Path, err)
		os.Exit(1)
	}

	file2Content, err := os.ReadFile(file2Path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", file2Path, err)
		os.Exit(1)
	}

	// Generate HTML
	html, err := GenerateHTML(string(file1Content), string(file2Content), fileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating diff: %v\n", err)
		os.Exit(1)
	}

	// Save debug copy if debug mode is enabled
	if debugMode {
		debugDir := "cdiff_debug_files"
		// Get absolute path for clarity
		absDebugDir, _ := filepath.Abs(debugDir)
		if err := os.MkdirAll(debugDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "cdiff debug: could not create directory %s: %v\n", absDebugDir, err)
		} else {
			// Use fileName with .html extension for the debug file
			debugFileName := fileName + ".html"
			debugPath := filepath.Join(debugDir, debugFileName)
			absDebugPath, _ := filepath.Abs(debugPath)
			if err := os.WriteFile(debugPath, []byte(html), 0644); err != nil {
				fmt.Fprintf(os.Stderr, "cdiff debug: could not write file %s: %v\n", absDebugPath, err)
			} else {
				fmt.Fprintf(os.Stderr, "cdiff debug: saved to %s\n", absDebugPath)
			}
		}
	}

	// Create file entry for IPC or new window
	entry := FileEntry{
		Name:    fileName,
		Path:    "", // Generated content, no file path
		Content: html,
	}

	// Try to send to existing instance first (sidebar mode)
	if TrySendToSidebarInstance(entry) {
		os.Exit(0)
	}

	// No existing instance - write HTML to temp file and spawn GUI
	tmpFile, err := os.CreateTemp("", "cdiff-*.html")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating temp file: %v\n", err)
		os.Exit(1)
	}
	if _, err := tmpFile.WriteString(html); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		fmt.Fprintf(os.Stderr, "Error writing temp file: %v\n", err)
		os.Exit(1)
	}
	tmpFile.Close()

	// Spawn GUI in background and exit
	if err := spawnGUIBackground(tmpFile.Name(), fileName); err != nil {
		os.Remove(tmpFile.Name())
		fmt.Fprintf(os.Stderr, "Error spawning GUI: %v\n", err)
		os.Exit(1)
	}

	os.Exit(0)
}

func printUsage() {
	fmt.Println("Usage: cdiff [--debug] <file1> <file2> [filename]")
	fmt.Println()
	fmt.Println("Arguments:")
	fmt.Println("  file1     Path to the first file (old version)")
	fmt.Println("  file2     Path to the second file (new version)")
	fmt.Println("  filename  Optional display name for the file being compared")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --debug        Save generated HTML to cdiff_debug_files/ directory")
	fmt.Println("  -v, --version  Show version")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  cdiff old.txt new.txt")
	fmt.Println("  cdiff --debug old.txt new.txt")
	fmt.Println("  cdiff /tmp/abc1234_file.go /tmp/xyz5678_file.go file.go")
	fmt.Println()
	fmt.Println("Git difftool configuration:")
	fmt.Println("  git config --global diff.tool cdiff")
	fmt.Println("  git config --global difftool.cdiff.cmd 'cdiff \"$LOCAL\" \"$REMOTE\" \"$MERGED\"'")
}

// spawnGUIBackground spawns the GUI as a background process
func spawnGUIBackground(htmlPath, displayName string) error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	args := []string{"--internal-gui", "--html", htmlPath, displayName}

	// Spawn the child process detached
	cmd := exec.Command(exe, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true, // Create new session so child survives parent exit
	}
	// Don't inherit stdin, but keep stderr for errors
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start GUI process: %w", err)
	}

	// Wait for socket to be created (guarantees subsequent invocations can connect)
	socketPath := getSidebarSocketPath()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(socketPath); err == nil {
			return nil // Socket exists, child is ready
		}
		time.Sleep(10 * time.Millisecond)
	}

	return fmt.Errorf("timeout waiting for GUI to start")
}

// runGUI runs the Wails application (called from GUI subprocess)
func runGUI(htmlPath, displayName string) {
	// Read the HTML content
	content, err := os.ReadFile(htmlPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading HTML file: %v\n", err)
		os.Exit(1)
	}

	// Clean up the temp file
	os.Remove(htmlPath)

	// Create file entry
	entry := FileEntry{
		Name:    displayName,
		Path:    "", // Generated content, no file path
		Content: string(content),
	}

	// Create app with the file entry
	app := NewApp(entry)

	// Start IPC server for sidebar mode
	ipcServer, err := StartSidebarServer(app)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not start IPC server: %v\n", err)
	}

	// Load saved window state
	state := LoadWindowState()
	config := app.config

	// Determine window dimensions
	width, height := GetWindowDimensions(state, config)

	// Determine window position (to be set after startup)
	x, y, shouldSetPosition := GetWindowPosition(state, config)
	app.initialX = x
	app.initialY = y
	app.initialWidth = width
	app.initialHeight = height
	app.shouldSetPosition = shouldSetPosition

	// Create local file handler for serving relative assets
	localFileHandler := NewLocalFileHandler(app)

	// Run Wails application
	err = wails.Run(&options.App{
		Title:     displayName,
		Width:     width,
		Height:    height,
		MinWidth:  MinWindowWidth,
		MinHeight: MinWindowHeight,
		AssetServer: &assetserver.Options{
			Assets:  assets,
			Handler: localFileHandler,
		},
		OnStartup: app.startup,
		OnShutdown: func(ctx context.Context) {
			if ipcServer != nil {
				ipcServer.Close()
			}
		},
		Bind: []interface{}{
			app,
		},
		Mac: &mac.Options{
			TitleBar:             mac.TitleBarDefault(),
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
		},
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
