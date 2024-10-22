package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPathValidate(t *testing.T) {
	tests := []struct {
		name    string
		path    Path
		wantErr bool
	}{
		{"Valid path with destination", Path{Path: "/api", Destination: ptr("http://example.com")}, false},
		{"Valid path with directory", Path{Path: "/static", Directory: ptr("/var")}, false},
		{"Invalid path without leading slash", Path{Path: "api", Destination: ptr("http://example.com")}, true},
		{"Invalid destination URL", Path{Path: "/api", Destination: ptr("not-a-url")}, true},
		{"Invalid with both destination and directory", Path{Path: "/both", Destination: ptr("http://example.com"), Directory: ptr("/var/www")}, true},
		{"Invalid with neither destination nor directory", Path{Path: "/empty"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.path.validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Path.validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestServiceValidate(t *testing.T) {
	tests := []struct {
		name    string
		service Service
		wantErr bool
	}{
		{"Valid service", Service{Domain: "example.com", Paths: []Path{{Path: "/api", Destination: ptr("http://api.example.com")}}}, false},
		{"Invalid domain", Service{Domain: "not a domain", Paths: []Path{{Path: "/api", Destination: ptr("http://api.example.com")}}}, true},
		{"Invalid path", Service{Domain: "example.com", Paths: []Path{{Path: "invalid", Destination: ptr("http://api.example.com")}}}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.service.validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Service.validate() service = %v, error = %v, wantErr %v", err, tt.service, tt.wantErr)
			}
		})
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name           string
		config         Config
		currentVersion string
		wantErr        bool
	}{
		{"Valid config", Config{Version: "1.0.0", Host: "localhost", Port: 80, Services: []Service{{Domain: "example.com", Paths: []Path{{Path: "/api", Destination: ptr("http://api.example.com")}}}}}, "1.0.0", false},
		{"AutoTLS with port != 443", Config{Version: "1.0.0", Host: "localhost", Port: 80, TLS: TLS{Auto: true, Domains: []string{"example.com"}}, Services: []Service{{Domain: "example.com", Paths: []Path{{Path: "/api", Destination: ptr("http://api.example.com")}}}}}, "1.0.0", true},
		{"Missing version", Config{Host: "localhost"}, "1.0.0", true},
		{"Invalid version", Config{Version: "invalid", Host: "localhost"}, "1.0.0", true},
		{"Future version", Config{Version: "2.0.0", Host: "localhost"}, "1.0.0", true},
		{"Missing host", Config{Version: "1.0.0"}, "1.0.0", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate(tt.currentVersion)
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseConfig(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "config_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a valid config file
	validConfig := `
version: "1.0.0"
host: "localhost"
port: 8080
services:
  - domain: "example.com"
    endpoints:
      - path: "/api"
        destination: "http://api.example.com"
`
	validConfigPath := filepath.Join(tempDir, "valid_config.yaml")
	err = os.WriteFile(validConfigPath, []byte(validConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to write valid config file: %v", err)
	}

	// Create an invalid config file
	invalidConfig := `
version: "invalid"
host: "localhost"
`
	invalidConfigPath := filepath.Join(tempDir, "invalid_config.yaml")
	err = os.WriteFile(invalidConfigPath, []byte(invalidConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to write invalid config file: %v", err)
	}

	tests := []struct {
		name           string
		filepath       string
		currentVersion string
		wantErr        bool
	}{
		{"Valid config", validConfigPath, "1.0.0", false},
		{"Invalid config", invalidConfigPath, "1.0.0", true},
		{"Non-existent file", filepath.Join(tempDir, "non_existent.yaml"), "1.0.0", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseConfig(tt.filepath, tt.currentVersion)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsValidURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{"Valid URL", "http://example.com", true},
		{"Valid URL with path", "https://example.com/path", true},
		{"Invalid URL", "not-a-url", false},
		{"Invalid URL", "not a domain", false},
		{"Missing scheme", "example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidURL(tt.url); got != tt.want {
				t.Errorf("isValidURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsValidDir(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "dir_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name string
		path string
		want bool
	}{
		{"Valid directory", tempDir, true},
		{"Non-existent directory", filepath.Join(tempDir, "non_existent"), false},
		{"Empty path", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidDir(tt.path); got != tt.want {
				t.Errorf("isValidDir() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Helper function to create string pointers
func ptr(s string) *string {
	return &s
}
