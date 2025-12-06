package calendar

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// calendarScopes defines the OAuth2 scopes required for Google Calendar operations.
var calendarScopes = []string{
	calendar.CalendarEventsScope, // Create and manage calendar events
}

// launchWebServer indicates whether to use web server auth flow (true) or manual code entry (false)
const launchWebServer = true

// Make exec.Command replaceable for testing
var execCommand = exec.Command

// CalendarService wraps Google Calendar API operations
type CalendarService struct {
	service *calendar.Service
}

// NewCalendarService creates a new CalendarService with authenticated client
func NewCalendarService(ctx context.Context) (*CalendarService, error) {
	client, err := getClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get calendar client: %w", err)
	}

	service, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create calendar service: %w", err)
	}

	return &CalendarService{service: service}, nil
}

// CreateVideoReleaseEvent creates a calendar event for a video release
// startTime is the video publish time, the event spans from 30 minutes before to 30 minutes after
func (cs *CalendarService) CreateVideoReleaseEvent(ctx context.Context, videoTitle, videoURL string, publishTime time.Time) (*calendar.Event, error) {
	// Calculate event times: 30 minutes before to 30 minutes after publish time
	eventStart := publishTime.Add(-30 * time.Minute)
	eventEnd := publishTime.Add(30 * time.Minute)

	event := &calendar.Event{
		Summary:     fmt.Sprintf("📺 Video Release: %s", videoTitle),
		Description: fmt.Sprintf("Video going live!\n\nYouTube URL: %s\n\nTasks:\n- Post on X (Twitter)\n- Monitor early comments\n- Engage with viewers\n- Share on additional platforms", videoURL),
		Start: &calendar.EventDateTime{
			DateTime: eventStart.Format(time.RFC3339),
			TimeZone: "UTC",
		},
		End: &calendar.EventDateTime{
			DateTime: eventEnd.Format(time.RFC3339),
			TimeZone: "UTC",
		},
		Reminders: &calendar.EventReminders{
			UseDefault: true,
		},
	}

	createdEvent, err := cs.service.Events.Insert("primary", event).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to create calendar event: %w", err)
	}

	return createdEvent, nil
}

// getClient uses a Context to retrieve a Token and generate a Client.
// It uses the centralized calendarScopes for all Calendar operations.
// If modifying the scopes, delete your previously saved credentials at ~/.credentials/calendar-go.json
func getClient(ctx context.Context) (*http.Client, error) {
	b, err := os.ReadFile("client_secret.json")
	if err != nil {
		return nil, fmt.Errorf("unable to read client secret file: %w", err)
	}

	config, err := google.ConfigFromJSON(b, calendarScopes...)
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file to config: %w", err)
	}

	// Use a redirect URI like this for a web app. The redirect URI must be a
	// valid one for your OAuth2 credentials.
	config.RedirectURL = "http://localhost:8090"

	cacheFile, err := tokenCacheFile()
	if err != nil {
		return nil, fmt.Errorf("unable to get path to cached credential file: %w", err)
	}

	tok, err := tokenFromFile(cacheFile)
	if err != nil {
		authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
		if launchWebServer {
			fmt.Println("Trying to get token from web for Google Calendar")
			tok, err = getTokenFromWeb(config, authURL)
		} else {
			fmt.Println("Trying to get token from prompt for Google Calendar")
			tok, err = getTokenFromPrompt(config, authURL)
		}
		if err != nil {
			return nil, fmt.Errorf("unable to get token: %w", err)
		}
		saveToken(cacheFile, tok)
	}
	return config.Client(ctx, tok), nil
}

// startWebServer starts a web server that listens on http://localhost:8090.
// The webserver waits for an oauth code in the three-legged auth flow.
func startWebServer() (codeCh chan string, err error) {
	listener, err := net.Listen("tcp", "localhost:8090")
	if err != nil {
		return nil, err
	}
	codeCh = make(chan string)

	go http.Serve(listener, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		code := r.FormValue("code")
		codeCh <- code // send code to OAuth flow
		listener.Close()
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "Received code: %v\r\nYou can now safely close this browser window.", code)
	}))

	return codeCh, nil
}

// openURL opens a browser window to the specified location.
func openURL(url string) error {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = execCommand("xdg-open", url).Start()
	case "windows":
		err = execCommand("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = execCommand("open", url).Start()
	default:
		err = fmt.Errorf("cannot open URL %s on this platform", url)
	}
	return err
}

// exchangeToken exchanges the authorization code for an access token
func exchangeToken(config *oauth2.Config, code string) (*oauth2.Token, error) {
	tok, err := config.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve token: %w", err)
	}
	return tok, nil
}

// getTokenFromPrompt uses Config to request a Token and prompts the user
// to enter the token on the command line. It returns the retrieved Token.
func getTokenFromPrompt(config *oauth2.Config, authURL string) (*oauth2.Token, error) {
	var code string
	fmt.Printf("Go to the following link in your browser. After completing "+
		"the authorization flow, enter the authorization code on the command "+
		"line: \n%v\n", authURL)

	if _, err := fmt.Scan(&code); err != nil {
		return nil, fmt.Errorf("unable to read authorization code: %w", err)
	}
	return exchangeToken(config, code)
}

// getTokenFromWeb uses Config to request a Token via web server.
// It returns the retrieved Token.
func getTokenFromWeb(config *oauth2.Config, authURL string) (*oauth2.Token, error) {
	codeCh, err := startWebServer()
	if err != nil {
		fmt.Printf("Unable to start a web server.")
		return nil, err
	}

	err = openURL(authURL)
	if err != nil {
		log.Printf("Unable to open authorization URL in web server: %v", err)
		// Fall back to manual prompt
		return getTokenFromPrompt(config, authURL)
	}

	fmt.Println("Your browser has been opened to an authorization URL.",
		" This program will resume once authorization has been provided.")
	fmt.Println(authURL)

	// Wait for the web server to get the code.
	code := <-codeCh
	return exchangeToken(config, code)
}

// tokenCacheFile generates credential file path/filename.
// It returns the generated credential path/filename.
func tokenCacheFile() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	tokenCacheDir := filepath.Join(usr.HomeDir, ".credentials")
	os.MkdirAll(tokenCacheDir, 0700)
	return filepath.Join(tokenCacheDir,
		url.QueryEscape("calendar-go.json")), err
}

// tokenFromFile retrieves a Token from a given file path.
// It returns the retrieved Token and any read error encountered.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	return t, err
}

// saveToken uses a file path to create a file and store the
// token in it.
func saveToken(file string, token *oauth2.Token) {
	fmt.Println("Saving Google Calendar credential")
	fmt.Printf("Saving credential file to: %s\n", file)
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Printf("Unable to cache oauth token: %v", err)
		return
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}
