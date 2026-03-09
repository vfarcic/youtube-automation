package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// OAuthConfig holds OAuth configuration for a Google API service.
type OAuthConfig struct {
	CredentialsFile string   // Path to client_secret JSON file
	TokenFileName   string   // Name of token cache file (stored in ~/.credentials/)
	CallbackPort    int      // Port for OAuth callback web server
	Scopes          []string // OAuth2 scopes to request
}

// ExecCommandFunc allows replacing exec.Command for testing.
var ExecCommandFunc = exec.Command

// GetClient builds an authenticated *http.Client from the given config.
// If a cached token exists, it is reused. Otherwise, the OAuth browser flow is launched.
func GetClient(ctx context.Context, cfg OAuthConfig) (*http.Client, error) {
	b, err := os.ReadFile(cfg.CredentialsFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read client secret file %s: %w", cfg.CredentialsFile, err)
	}

	config, err := google.ConfigFromJSON(b, cfg.Scopes...)
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file to config: %w", err)
	}

	config.RedirectURL = fmt.Sprintf("http://localhost:%d", cfg.CallbackPort)

	cacheFile, err := TokenCacheFileWithName(cfg.TokenFileName)
	if err != nil {
		return nil, fmt.Errorf("unable to get path to cached credential file: %w", err)
	}

	tok, err := TokenFromFile(cacheFile)
	if err != nil {
		authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
		fmt.Println("Trying to get token from web")
		tok, err = getTokenFromWebWithPort(config, authURL, cfg.CallbackPort)
		if err != nil {
			return nil, fmt.Errorf("unable to get token: %w", err)
		}
		if saveErr := SaveToken(cacheFile, tok); saveErr != nil {
			return nil, fmt.Errorf("unable to cache oauth token: %w", saveErr)
		}
	}
	return config.Client(ctx, tok), nil
}

// TokenCacheFileWithName generates credential file path with the specified filename.
// Tokens are stored under ~/.credentials/.
func TokenCacheFileWithName(filename string) (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	tokenCacheDir := filepath.Join(usr.HomeDir, ".credentials")
	if _, statErr := os.Stat(tokenCacheDir); os.IsNotExist(statErr) {
		if err := os.MkdirAll(tokenCacheDir, 0700); err != nil {
			return "", fmt.Errorf("failed to create token cache directory: %w", err)
		}
	}
	return filepath.Join(tokenCacheDir, url.QueryEscape(filename)), nil
}

// TokenFromFile retrieves a Token from a given file path.
func TokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	return t, err
}

// SaveToken writes the token to the given file path.
func SaveToken(file string, token *oauth2.Token) error {
	fmt.Println("trying to save token")
	fmt.Printf("Saving credential file to: %s\n", file)
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("unable to cache oauth token: %w", err)
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(token)
}

// startWebServerWithPort starts a web server on the specified port.
// The webserver waits for an oauth code in the three-legged auth flow.
func startWebServerWithPort(port int) (codeCh chan string, err error) {
	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		return nil, err
	}
	codeCh = make(chan string)

	go http.Serve(listener, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		code := r.FormValue("code")
		codeCh <- code
		listener.Close()
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "Received code: %v\r\nYou can now safely close this browser window.", code)
	}))

	return codeCh, nil
}

// openURL opens a browser window to the specified location.
func openURL(urlStr string) error {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = ExecCommandFunc("xdg-open", urlStr).Start()
	case "windows":
		err = ExecCommandFunc("rundll32", "url.dll,FileProtocolHandler", urlStr).Start()
	case "darwin":
		err = ExecCommandFunc("open", urlStr).Start()
	default:
		err = fmt.Errorf("cannot open URL %s on this platform", urlStr)
	}
	return err
}

// exchangeToken exchanges the authorization code for an access token.
func exchangeToken(config *oauth2.Config, code string) (*oauth2.Token, error) {
	tok, err := config.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve token: %w", err)
	}
	return tok, nil
}

// getTokenFromWebWithPort uses Config to request a Token using a specific port.
func getTokenFromWebWithPort(config *oauth2.Config, authURL string, port int) (*oauth2.Token, error) {
	codeCh, err := startWebServerWithPort(port)
	if err != nil {
		return nil, fmt.Errorf("unable to start web server on port %d: %w", port, err)
	}

	err = openURL(authURL)
	if err != nil {
		return nil, fmt.Errorf("unable to open authorization URL in browser: %w", err)
	}
	fmt.Println("Your browser has been opened to an authorization URL.",
		" This program will resume once authorization has been provided.")
	fmt.Println(authURL)

	code := <-codeCh
	return exchangeToken(config, code)
}
