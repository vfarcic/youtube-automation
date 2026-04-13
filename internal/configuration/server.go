package configuration

import (
	"os"
	"strconv"
)

const (
	DefaultHost    = "localhost"
	DefaultPort    = 8080
	DefaultDataDir = "./tmp"
)

// GetServeHost returns the configured host for the API server.
// Priority: SERVER_HOST env var > default
func GetServeHost() string {
	if host := os.Getenv("SERVER_HOST"); host != "" {
		return host
	}
	return DefaultHost
}

// GetServePort returns the configured port for the API server.
// Priority: SERVER_PORT env var > default
func GetServePort() int {
	if portStr := os.Getenv("SERVER_PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			return port
		}
	}
	return DefaultPort
}

// GetAPIToken returns the configured API authentication token.
// Priority: API_TOKEN env var > settings.yaml api.token
func GetAPIToken() string {
	if token := os.Getenv("API_TOKEN"); token != "" {
		return token
	}
	return GlobalSettings.API.Token
}

// GetDataDir returns the configured data directory for video YAML files.
// Priority: DATA_DIR env var > default
func GetDataDir() string {
	if dir := os.Getenv("DATA_DIR"); dir != "" {
		return dir
	}
	return DefaultDataDir
}
