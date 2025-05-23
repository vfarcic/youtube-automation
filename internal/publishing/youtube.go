package publishing

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
	"strings"

	"devopstoolkit/youtube-automation/internal/configuration"
	"devopstoolkit/youtube-automation/internal/storage"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
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
func getClient(ctx context.Context, config *oauth2.Config) *http.Client {
	b, err := os.ReadFile("client_secret.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying the scope, delete your previously saved credentials
	// at ~/.credentials/youtube-go.json
	configFromJSON, err := google.ConfigFromJSON(b, config.Scopes...)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	config = configFromJSON

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

// Make exec.Command replaceable for testing
var execCommand = exec.Command

// openURL opens a browser window to the specified location.
// This code originally appeared at:
//
//	http://stackoverflow.com/questions/10377243/how-can-i-launch-a-process-that-is-not-a-file-in-go
func openURL(url string) error {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = execCommand("xdg-open", url).Start()
	case "windows":
		err = execCommand("rundll32", "url.dll,FileProtocolHandler", "http://localhost:4001/").Start()
	case "darwin":
		err = execCommand("open", url).Start()
	default:
		err = fmt.Errorf("cannot open URL %s on this platform", url)
	}
	return err
}

// Exchange the authorization code for an access token
func exchangeToken(config *oauth2.Config, code string) (*oauth2.Token, error) {
	tok, err := config.Exchange(context.Background(), code)
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

func UploadVideo(video *storage.Video) string {
	if video.UploadVideo == "" {
		log.Fatalf("You must provide a filename of a video file to upload")
		return ""
	}
	if video.Thumbnail == "" {
		log.Fatalf("You must provide a thumbnail of the video file to upload")
		return ""
	}
	client := getClient(context.Background(), &oauth2.Config{Scopes: []string{youtube.YoutubeUploadScope}})

	// FIXME: Remove the comment
	// service, err := youtube.New(client)
	ctx := context.Background()
	service, err := youtube.NewService(ctx, option.WithHTTPClient(client))
	// service, err := youtube.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Error creating YouTube client: %v", err)
	}
	timecodes := ""
	if len(video.Timecodes) > 0 && video.Timecodes != "N/A" {
		timecodes = fmt.Sprintf("▬▬▬▬▬▬ ⏱ Timecodes ⏱ ▬▬▬▬▬▬\n%s", video.Timecodes)
	}
	description := fmt.Sprintf(`%s

%s

Consider joining the channel: https://www.youtube.com/c/devopstoolkit/join

▬▬▬▬▬▬ 🔗 Additional Info 🔗 ▬▬▬▬▬▬ 
%s
▬▬▬▬▬▬ 💰 Sponsorships 💰 ▬▬▬▬▬▬ 
If you are interested in sponsoring this channel, please visit https://devopstoolkit.live/sponsor for more information. Alternatively, feel free to contact me over Twitter or LinkedIn (see below).

▬▬▬▬▬▬ 👋 Contact me 👋 ▬▬▬▬▬▬ 
➡ BlueSky: https://vfarcic.bsky.social
➡ LinkedIn: https://www.linkedin.com/in/viktorfarcic/

▬▬▬▬▬▬ 🚀 Other Channels 🚀 ▬▬▬▬▬▬
🎤 Podcast: https://www.devopsparadox.com/
💬 Live streams: https://www.youtube.com/c/DevOpsParadox

%s
`, video.Description, video.DescriptionTags, GetAdditionalInfo(video.HugoPath, video.ProjectName, video.ProjectURL, video.RelatedVideos), timecodes)

	upload := &youtube.Video{
		Snippet: &youtube.VideoSnippet{
			Title:       video.Title,
			Description: description,
			CategoryId:  "28",
			ChannelId:   channelID,
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

	// Determine languages to set
	finalDefaultLanguage := video.Language
	if finalDefaultLanguage == "" {
		finalDefaultLanguage = configuration.GlobalSettings.VideoDefaults.Language // Guaranteed non-empty by cli.go
	}

	finalDefaultAudioLanguage := video.AudioLanguage
	if finalDefaultAudioLanguage == "" {
		finalDefaultAudioLanguage = configuration.GlobalSettings.VideoDefaults.AudioLanguage // Guaranteed non-empty by cli.go
	}

	upload.Snippet.DefaultLanguage = finalDefaultLanguage
	upload.Snippet.DefaultAudioLanguage = finalDefaultAudioLanguage

	call := service.Videos.Insert([]string{"snippet", "status"}, upload)
	file, err := os.Open(video.UploadVideo)
	if err != nil {
		log.Fatalf("Error opening %v: %v", video.UploadVideo, err)
	}

	response, err := call.Media(file).Do()
	file.Close()
	if err != nil {
		log.Fatalf("Error getting response from YouTube during insert: %v", err)
	}
	fmt.Printf("Upload successful! Video ID: %v\n", response.Id)

	// Save the applied languages back to the video struct
	video.AppliedLanguage = finalDefaultLanguage
	video.AppliedAudioLanguage = finalDefaultAudioLanguage
	log.Printf("DEBUG: Language %s and Audio Language %s stored in video struct for video ID %s", video.AppliedLanguage, video.AppliedAudioLanguage, response.Id)

	adapter := &youtubeServiceAdapter{service: service} // Create adapter
	err = updateVideoLanguage(adapter, response.Id, finalDefaultLanguage, finalDefaultAudioLanguage)
	if err != nil {
		log.Printf("Error updating video languages for video ID %s: %v", response.Id, err)
	} else {
		fmt.Printf("Successfully set language to %s and audio language to %s for video ID %s\n", finalDefaultLanguage, finalDefaultAudioLanguage, response.Id)
	}

	return response.Id
}

func GetAdditionalInfo(hugoPath, projectName, projectURL, relatedVideosRaw string) string {
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
	if len(hugoPath) > 0 {
		hugoPage := strings.ReplaceAll(hugoPath, "../", "")
		hugoPage = strings.ReplaceAll(hugoPage, "devopstoolkit-live/content/", "")
		hugoPage = strings.ReplaceAll(hugoPage, "/_index.md", "")
		hugoUrl := fmt.Sprintf("https://devopstoolkit.live/%s", hugoPage)
		gist = fmt.Sprintf("➡ Transcript and commands: %s\n", hugoUrl)
	}
	return fmt.Sprintf("%s🔗 %s: %s\n%s", gist, projectName, projectURL, relatedVideos)
}

func UploadThumbnail(video storage.Video) error {
	client := getClient(context.Background(), &oauth2.Config{Scopes: []string{youtube.YoutubeUploadScope}})

	// FIXME: Remove the comment
	// service, err := youtube.New(client)
	ctx := context.Background()
	service, err := youtube.NewService(ctx, option.WithHTTPClient(client))
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

func GetYouTubeURL(videoId string) string {
	return fmt.Sprintf("https://youtu.be/%s", videoId)
}

// videoUpdateDoer defines an interface for the Do() method of a video update call.
type videoUpdateDoer interface {
	Do(opts ...googleapi.CallOption) (*youtube.Video, error)
}

// videoServiceUpdater defines an interface for the Update() method of a video service.
type videoServiceUpdater interface {
	Update(part []string, video *youtube.Video) videoUpdateDoer
}

// youtubeServiceAdapter adapts *youtube.Service to the videoServiceUpdater interface.
type youtubeServiceAdapter struct {
	service *youtube.Service
}

// Update calls the underlying YouTube service's Videos.Update method.
func (a *youtubeServiceAdapter) Update(part []string, video *youtube.Video) videoUpdateDoer {
	return a.service.Videos.Update(part, video)
}

func updateVideoLanguage(updater videoServiceUpdater, videoID string, languageCode string, audioLanguageCode string) error {
	// Determine final language codes with fallbacks
	finalLangCode := languageCode
	if finalLangCode == "" {
		finalLangCode = configuration.GlobalSettings.VideoDefaults.Language // Guaranteed non-empty by cli.go
	}

	finalAudioLangCode := audioLanguageCode
	if finalAudioLangCode == "" {
		finalAudioLangCode = configuration.GlobalSettings.VideoDefaults.AudioLanguage // Guaranteed non-empty by cli.go
	}

	updateVideo := &youtube.Video{
		Id: videoID,
		Snippet: &youtube.VideoSnippet{
			DefaultLanguage:      finalLangCode,
			DefaultAudioLanguage: finalAudioLangCode,
		},
	}

	updateCall := updater.Update([]string{"snippet"}, updateVideo)
	_, err := updateCall.Do()
	return err
}
