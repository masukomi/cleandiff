package main

import (
	"os"
	"path/filepath"
)

// Config holds the application configuration
type Config struct {
	// FontSize is the default font size in pixels (e.g., 16, 18, 24)
	FontSize int `json:"font_size"`
	// ChromeCSS is the path to a custom CSS file for styling cleandiff UI
	ChromeCSS string `json:"chrome_css"`
	// DefaultWidth is the default window width in pixels (0 = use app default)
	DefaultWidth int `json:"default_width"`
	// DefaultHeight is the default window height in pixels (0 = use app default)
	DefaultHeight int `json:"default_height"`
	// DefaultX is the default window X position in pixels (0 = use system default)
	DefaultX int `json:"default_x"`
	// DefaultY is the default window Y position in pixels (0 = use system default)
	DefaultY int `json:"default_y"`
}

// DefaultConfig returns the default configuration values
func DefaultConfig() Config {
	return Config{
		FontSize: 0, // 0 means use browser default
	}
}

// getConfigDir returns the config directory following XDG Base Directory standard
func getConfigDir() string {
	// Check XDG_CONFIG_HOME first
	if xdgConfigHome := os.Getenv("XDG_CONFIG_HOME"); xdgConfigHome != "" {
		return filepath.Join(xdgConfigHome, "cleandiff")
	}
	// Fall back to ~/.config/cleandiff
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "cleandiff")
}

// LoadConfig returns the default configuration
// (cleandiff doesn't use a config file, just template overrides)
func LoadConfig() Config {
	return DefaultConfig()
}
