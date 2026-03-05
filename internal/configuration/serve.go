package configuration

import (
	"os"

	"github.com/spf13/cobra"
)

var serveHost string
var servePort int
var serveMode bool
var apiToken string

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the HTTP API server",
	Run: func(cmd *cobra.Command, args []string) {
		serveMode = true
	},
}

func init() {
	serveCmd.Flags().StringVar(&serveHost, "host", "localhost", "Host to listen on")
	serveCmd.Flags().IntVar(&servePort, "port", 8080, "Port to listen on")
	serveCmd.Flags().StringVar(&apiToken, "api-token", os.Getenv("API_TOKEN"), "Bearer token for API authentication (or set API_TOKEN env var)")
	RootCmd.AddCommand(serveCmd)
}

// GetServeHost returns the configured host for the API server.
func GetServeHost() string {
	return serveHost
}

// GetServePort returns the configured port for the API server.
func GetServePort() int {
	return servePort
}

// GetAPIToken returns the configured API authentication token.
// Priority: --api-token flag > API_TOKEN env var > settings.yaml api.token
func GetAPIToken() string {
	if apiToken != "" {
		return apiToken
	}
	return GlobalSettings.API.Token
}

// IsServeMode returns true when the serve subcommand was invoked.
func IsServeMode() bool {
	return serveMode
}
