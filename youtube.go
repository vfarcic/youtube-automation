package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/youtube/v3"
)

const channelID = "UCfz8x0lVzJpb_dgWm9kPVrw"

// This variable indicates whether the script should launch a web server to
// initiate the authorization flow or just display the URL in the terminal
// window. Note the following instructions based on this setting:
// * launchWebServer = true
//  1. Use OAuth2 credentials for a web application
//  2. Define authorized redirect URIs for the credential in the Google APIs
//     Console and set the RedirectURL property on the config object to one
//     of those redirect URIs. For example:
//     config.RedirectURL = "http://localhost:8090"
//  3. In the startWebServer function below, update the URL in this line
//     to match the redirect URI you selected:
//     listener, err := net.Listen("tcp", "localhost:8090")
//     The redirect URI identifies the URI to which the user is sent after
//     completing the authorization flow. The listener then captures the
//     authorization code in the URL and passes it back to this script.
//
// * launchWebServer = false
//  1. Use OAuth2 credentials for an installed application. (When choosing
//     the application type for the OAuth2 client ID, select "Other".)
//  2. Set the redirect URI to "urn:ietf:wg:oauth:2.0:oob", like this:
//     config.RedirectURL = "urn:ietf:wg:oauth:2.0:oob"
//  3. When running the script, complete the auth flow. Then copy the
//     authorization code from the browser and enter it on the command line.
const launchWebServer = true

// TODO: Remove
// const missingClientSecretsMessage = `
// Please configure OAuth 2.0
// To make this sample run, you need to populate the client_secrets.json file
// found at:
//    %v
// with information from the {{ Google Cloud Console }}
// {{ https://cloud.google.com/console }}
// For more information about the client_secrets.json file format, please visit:
// https://developers.google.com/api-client-library/python/guide/aaa_client_secrets
// `

// getClient uses a Context and Config to retrieve a Token
// then generate a Client. It returns the generated Client.
func getClient(scope string) *http.Client {
	ctx := context.Background()

	b, err := os.ReadFile("client_secret.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying the scope, delete your previously saved credentials
	// at ~/.credentials/youtube-go.json
	config, err := google.ConfigFromJSON(b, scope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	// Use a redirect URI like this for a web app. The redirect URI must be a
	// valid one for your OAuth2 credentials.
	config.RedirectURL = "http://localhost:8090"
	// Use the following redirect URI if launchWebServer=false in oauth2.go
	// config.RedirectURL = "urn:ietf:wg:oauth:2.0:oob"

	cacheFile, err := tokenCacheFile()
	if err != nil {
		log.Fatalf("Unable to get path to cached credential file. %v", err)
	}
	tok, err := tokenFromFile(cacheFile)
	if err != nil {
		authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
		if launchWebServer {
			fmt.Println("Trying to get token from web")
			tok, err = getTokenFromWeb(config, authURL)
		} else {
			fmt.Println("Trying to get token from prompt")
			tok, err = getTokenFromPrompt(config, authURL)
		}
		if err == nil {
			saveToken(cacheFile, tok)
		}
	}
	return config.Client(ctx, tok)
}

// startWebServer starts a web server that listens on http://localhost:8080.
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
// This code originally appeared at:
//
//	http://stackoverflow.com/questions/10377243/how-can-i-launch-a-process-that-is-not-a-file-in-go
func openURL(url string) error {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", "http://localhost:4001/").Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("cannot open URL %s on this platform", url)
	}
	return err
}

// Exchange the authorization code for an access token
func exchangeToken(config *oauth2.Config, code string) (*oauth2.Token, error) {
	tok, err := config.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Fatalf("Unable to retrieve token %v", err)
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
		log.Fatalf("Unable to read authorization code %v", err)
	}
	fmt.Println(authURL)
	return exchangeToken(config, code)
}

// getTokenFromWeb uses Config to request a Token.
// It returns the retrieved Token.
func getTokenFromWeb(config *oauth2.Config, authURL string) (*oauth2.Token, error) {
	codeCh, err := startWebServer()
	if err != nil {
		fmt.Printf("Unable to start a web server.")
		return nil, err
	}

	err = openURL(authURL)
	if err != nil {
		log.Fatalf("Unable to open authorization URL in web server: %v", err)
	} else {
		fmt.Println("Your browser has been opened to an authorization URL.",
			" This program will resume once authorization has been provided.")
		fmt.Println(authURL)
	}

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
		url.QueryEscape("youtube-go.json")), err
}

// tokenFromFile retrieves a Token from a given file path.
// It returns the retrieved Token and any read error encountered.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	defer f.Close()
	return t, err
}

// saveToken uses a file path to create a file and store the
// token in it.
func saveToken(file string, token *oauth2.Token) {
	fmt.Println("trying to save token")
	fmt.Printf("Saving credential file to: %s\n", file)
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func uploadVideo(video Video) string {
	if video.UploadVideo == "" {
		log.Fatalf("You must provide a filename of a video file to upload")
		return ""
	}
	if video.Thumbnail == "" {
		log.Fatalf("You must provide a thumbnail of the video file to upload")
		return ""
	}
	client := getClient(youtube.YoutubeUploadScope)
	service, err := youtube.New(client)
	if err != nil {
		log.Fatalf("Error creating YouTube client: %v", err)
	}
	description := fmt.Sprintf(`%s

%s

Consider joining the channel: https://www.youtube.com/c/devopstoolkit/join

▬▬▬▬▬▬ 🔗 Additional Info 🔗 ▬▬▬▬▬▬ 
%s
▬▬▬▬▬▬ 💰 Sponsoships 💰 ▬▬▬▬▬▬ 
If you are interested in sponsoring this channel, please use https://calendar.app.google/Q9eaDUHN8ibWBaA7A to book a timeslot that suits you, and we'll go over the details. Or feel free to contact me over Twitter or LinkedIn (see below).

▬▬▬▬▬▬ 👋 Contact me 👋 ▬▬▬▬▬▬ 
➡ Twitter: https://twitter.com/vfarcic
➡ LinkedIn: https://www.linkedin.com/in/viktorfarcic/

▬▬▬▬▬▬ 🚀 Other Channels 🚀 ▬▬▬▬▬▬
🎤 Podcast: https://www.devopsparadox.com/
💬 Live streams: https://www.youtube.com/c/DevOpsParadox

▬▬▬▬▬▬ ⏱ Timecodes ⏱ ▬▬▬▬▬▬
%s
`, video.Description, video.DescriptionTags, getAdditionalInfo(video.GistUrl, video.ProjectName, video.ProjectURL, video.RelatedVideos), video.Timecodes)

	upload := &youtube.Video{
		Snippet: &youtube.VideoSnippet{
			Title:       video.Title,
			Description: description,
			CategoryId:  "28",
		},
		Status: &youtube.VideoStatus{
			PrivacyStatus: "private",
			PublishAt:     video.Date,
		},
		// MonetizationDetails: &youtube.VideoMonetizationDetails{
		// 	Access: &youtube.AccessPolicy{
		// 		Allowed: true,
		// 	},
		// },
	}
	// The API returns a 400 Bad Request response if tags is an empty string.
	if strings.Trim(video.Tags, "") != "" {
		upload.Snippet.Tags = strings.Split(video.Tags, ",")
	}
	call := service.Videos.Insert([]string{"snippet", "status"}, upload)
	file, err := os.Open(video.UploadVideo)
	defer file.Close()
	if err != nil {
		log.Fatalf("Error opening %v: %v", video.UploadVideo, err)
	}

	response, err := call.Media(file).Do()
	if err != nil {
		log.Fatalf("Error getting response from YouTube: %v", err)
	}
	fmt.Printf("Upload successful! Video ID: %v\n", response.Id)
	return response.Id
}

func getAdditionalInfo(gistUrl, projectName, projectURL, relatedVideosRaw string) string {
	relatedVideos := ""
	relatedVideosArray := strings.Split(relatedVideosRaw, "\n")
	for i := range relatedVideosArray {
		relatedVideosArray[i] = strings.TrimSpace(relatedVideosArray[i])
	}
	for i := range relatedVideosArray {
		if len(relatedVideosArray[i]) > 0 && relatedVideosArray[i] != "N/A" {
			relatedVideos = fmt.Sprintf("%s🎬 %s\n", relatedVideos, relatedVideosArray[i])
		}
	}
	gist := ""
	if len(gistUrl) > 0 {
		gist = fmt.Sprintf("➡ Gist with the commands: %s\n", gistUrl)
	}
	return fmt.Sprintf("%s🔗 %s: %s\n%s", gist, projectName, projectURL, relatedVideos)
}

func uploadThumbnail(video Video) error {
	client := getClient(youtube.YoutubeUploadScope)

	service, err := youtube.New(client)
	if err != nil {
		return err
	}
	file, err := os.Open(video.Thumbnail)
	if err != nil {
		return err
	}
	defer file.Close()
	call := service.Thumbnails.Set(video.VideoId)
	response, err := call.Media(file).Do()
	if err != nil {
		return err
	}
	fmt.Printf("Thumbnail uploaded, URL: %s\n", response.Items[0].Default.Url)
	return nil
}

func setPlaylists(video Video) error {
	client := getClient(youtube.YoutubeScope)
	service, err := youtube.New(client)
	if err != nil {
		return err
	}
	call := service.PlaylistItems.Insert([]string{"snippet", "status"}, &youtube.PlaylistItem{
		Snippet: &youtube.PlaylistItemSnippet{
			PlaylistId: "PLyicRj904Z99X4rm7NFnZGD80aDK83Miv",
			ResourceId: &youtube.ResourceId{
				Kind:    "youtube#video",
				VideoId: video.VideoId,
			},
		},
	})
	_, err = call.Do()
	if err != nil {
		return err
	}
	return nil
}

type PlaylistListResponse struct {
	Items []struct {
		ID      string `json:"id"`
		Snippet struct {
			Title       string `json:"title"`
			Description string `json:"description"`
		} `json:"snippet"`
	} `json:"items"`
}

func getYouTubePlaylists() map[string]string {
	apiKey := os.Getenv("YOUTUBE_API_KEY")
	apiUrl := fmt.Sprintf("https://www.googleapis.com/youtube/v3/playlists?part=snippet&channelId=%s&key=%s&maxResults=50", channelID, apiKey)

	resp, err := http.Get(apiUrl)
	if err != nil {
		log.Fatalf("Error getting response from YouTube: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error getting response body from YouTube: %v", err)
	}

	var response PlaylistListResponse
	if err := json.Unmarshal(body, &response); err != nil {
		log.Fatalf("Error unmarshalling response body from YouTube: %v", err)
	}

	playlistTitles := make(map[string]string)
	for _, item := range response.Items {
		playlistTitles[item.ID] = item.Snippet.Title + " - " + item.ID
	}
	return playlistTitles
}

func getYouTubeURL(videoId string) string {
	return fmt.Sprintf("https://youtu.be/%s", videoId)
}
