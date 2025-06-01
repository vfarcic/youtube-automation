// Package testutil provides helper functions and utilities for testing
// YouTube Automation components without relying on external dependencies.
package testutil

import (
	"bytes"
	"devopstoolkit/youtube-automation/internal/storage"
	"devopstoolkit/youtube-automation/internal/video"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// Tester is an interface for test assertion functions
// It allows test helpers to work with both *testing.T and our mock implementations
type Tester interface {
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Helper()
}

// ---- Environment Setup/Teardown ----

// SetupTestDir creates a temporary directory for tests
func SetupTestDir(t testing.TB) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "youtube-automation-test-")
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	return dir
}

// TeardownTestDir removes a temporary test directory
func TeardownTestDir(t testing.TB, dir string) {
	t.Helper()
	err := os.RemoveAll(dir)
	if err != nil {
		t.Fatalf("Failed to remove test directory: %v", err)
	}
}

// CopyFile copies a file from src to dst
func CopyFile(t testing.TB, src, dst string) {
	t.Helper()
	input, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("Failed to read source file: %v", err)
	}

	err = os.WriteFile(dst, input, 0644)
	if err != nil {
		t.Fatalf("Failed to write destination file: %v", err)
	}
}

// WriteTestFile writes content to a file in the given test directory
func WriteTestFile(t testing.TB, dir, filename, content string) string {
	t.Helper()
	path := filepath.Join(dir, filename)
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	return path
}

// ---- Fixture Generation ----

// CreateYAMLFixture creates a YAML fixture file in the test directory
func CreateYAMLFixture(t testing.TB, dir, filename, content string) string {
	t.Helper()
	return WriteTestFile(t, dir, filename, content)
}

// CreateJSONFixture creates a JSON fixture file in the test directory
func CreateJSONFixture(t testing.TB, dir, filename string, data interface{}) string {
	t.Helper()
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal JSON data: %v", err)
	}
	return WriteTestFile(t, dir, filename, string(jsonData))
}

// ---- HTTP Test Helpers ----

// MockHTTPServer creates a test server with the given handler and returns its URL
func MockHTTPServer(t testing.TB, handler http.Handler) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(func() {
		server.Close()
	})
	return server
}

// MockHTTPResponse creates a mock HTTP response for testing
func MockHTTPResponse(t testing.TB, statusCode int, body string) *http.Response {
	t.Helper()
	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
}

// MockJSONResponse creates a mock HTTP response with JSON content
func MockJSONResponse(t testing.TB, statusCode int, data interface{}) *http.Response {
	t.Helper()
	jsonData, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("Failed to marshal JSON data: %v", err)
	}

	response := MockHTTPResponse(t, statusCode, string(jsonData))
	response.Header.Set("Content-Type", "application/json")
	return response
}

// ---- Output Validation Helpers ----

// CaptureOutput captures stdout output during function execution
func CaptureOutput(t testing.TB, fn func()) string {
	t.Helper()

	// Save and restore original stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	os.Stdout = w

	// Create a channel for error reporting from goroutine
	errCh := make(chan error, 1)
	outCh := make(chan string, 1)

	// Capture output in a separate goroutine
	go func() {
		var buf bytes.Buffer
		_, err := io.Copy(&buf, r)
		if err != nil {
			// Send error through channel instead of directly calling t.Errorf
			errCh <- fmt.Errorf("Failed to copy output: %v", err)
		} else {
			errCh <- nil
		}
		outCh <- buf.String()
	}()

	fn()

	// Restore stdout and get captured output
	w.Close()
	os.Stdout = oldStdout

	// Check for errors from the goroutine
	if err := <-errCh; err != nil {
		t.Errorf("%v", err)
	}

	return <-outCh
}

// ProvideInput simulates user input for interactive CLI testing
func ProvideInput(t testing.TB, input string, fn func()) {
	t.Helper()

	// Save and restore original stdin
	oldStdin := os.Stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	os.Stdin = r

	// Create a channel for error reporting from goroutine
	errCh := make(chan error, 1)

	// Write input in a separate goroutine
	go func() {
		_, err := w.Write([]byte(input))
		if err != nil {
			// Send error through channel instead of directly calling t.Errorf
			errCh <- fmt.Errorf("Failed to write input: %v", err)
		} else {
			errCh <- nil
		}
		w.Close()
	}()

	fn()

	// Restore stdin
	os.Stdin = oldStdin

	// Check for errors from the goroutine
	if err := <-errCh; err != nil {
		t.Errorf("%v", err)
	}
}

// ---- Structure Comparison Helpers ----

// AssertEqual compares expected and actual values and reports an error if they don't match
func AssertEqual(t Tester, expected, actual interface{}, msgAndArgs ...interface{}) {
	t.Helper()

	if !reflect.DeepEqual(expected, actual) {
		message := fmt.Sprintf("Expected: %v\nActual: %v", expected, actual)
		if len(msgAndArgs) > 0 {
			message = fmt.Sprintf("%s\n%s", fmt.Sprint(msgAndArgs...), message)
		}
		t.Errorf(message)
	}
}

// AssertNotEqual checks that expected and actual values are not equal
func AssertNotEqual(t Tester, expected, actual interface{}, msgAndArgs ...interface{}) {
	t.Helper()

	if reflect.DeepEqual(expected, actual) {
		message := fmt.Sprintf("Expected values to differ, but both equal: %v", expected)
		if len(msgAndArgs) > 0 {
			message = fmt.Sprintf("%s\n%s", fmt.Sprint(msgAndArgs...), message)
		}
		t.Errorf(message)
	}
}

// AssertTrue checks that the condition is true
func AssertTrue(t Tester, condition bool, msgAndArgs ...interface{}) {
	t.Helper()

	if !condition {
		message := "Expected condition to be true"
		if len(msgAndArgs) > 0 {
			message = fmt.Sprintf("%s\n%s", fmt.Sprint(msgAndArgs...), message)
		}
		t.Errorf(message)
	}
}

// AssertFalse checks that the condition is false
func AssertFalse(t Tester, condition bool, msgAndArgs ...interface{}) {
	t.Helper()

	if condition {
		message := "Expected condition to be false"
		if len(msgAndArgs) > 0 {
			message = fmt.Sprintf("%s\n%s", fmt.Sprint(msgAndArgs...), message)
		}
		t.Errorf(message)
	}
}

// AssertContains checks that the string contains the specified substring
func AssertContains(t Tester, str, substr string, msgAndArgs ...interface{}) {
	t.Helper()

	if !strings.Contains(str, substr) {
		message := fmt.Sprintf("Expected string %q to contain %q", str, substr)
		if len(msgAndArgs) > 0 {
			message = fmt.Sprintf("%s\n%s", fmt.Sprint(msgAndArgs...), message)
		}
		t.Errorf(message)
	}
}

// CompareYAML compares two YAML strings for semantic equality
func CompareYAML(t Tester, expected, actual string) {
	t.Helper()
	var expectedData, actualData interface{}

	err := yaml.Unmarshal([]byte(expected), &expectedData)
	if err != nil {
		t.Fatalf("Failed to parse expected YAML: %v", err)
	}

	err = yaml.Unmarshal([]byte(actual), &actualData)
	if err != nil {
		t.Fatalf("Failed to parse actual YAML: %v", err)
	}

	AssertEqual(t, expectedData, actualData, "YAML comparison")
}

// ---- YAML Fixture Loaders ----

// Video represents a video definition from yaml.go
type Video struct {
	Name                string
	Index               int
	Path                string
	Category            string
	Init                Tasks
	Work                Tasks
	Define              Tasks
	Edit                Tasks
	Publish             Tasks
	ProjectName         string
	ProjectURL          string
	Sponsorship         Sponsorship
	Date                string
	Delayed             bool
	Code                bool
	Screen              bool
	Head                bool
	Thumbnails          bool
	Diagrams            bool
	Title               string
	Description         string
	Highlight           string
	Tags                string
	DescriptionTags     string
	Location            string
	Tagline             string
	TaglineIdeas        string
	OtherLogos          string
	Screenshots         bool
	RequestThumbnail    bool
	Thumbnail           string
	Members             string
	Animations          string
	RequestEdit         bool
	Movie               bool
	Timecodes           string
	Gist                string
	HugoPath            string
	RelatedVideos       string
	UploadVideo         string
	VideoId             string
	Tweet               string
	LinkedInPosted      bool
	SlackPosted         bool
	HNPosted            bool
	DOTPosted           bool
	BlueSkyPosted       bool
	YouTubeHighlight    bool
	YouTubeComment      bool
	YouTubeCommentReply bool
	Slides              bool
	GDE                 bool
	Repo                string
	NotifiedSponsors    bool
}

// Tasks represents the task completion structure
type Tasks struct {
	Completed int
	Total     int
}

// Sponsorship represents sponsorship information
type Sponsorship struct {
	Amount  string
	Emails  string
	Blocked string
}

// VideoIndex represents an entry in the video index
type VideoIndex struct {
	Name     string
	Category string
}

// Config represents the application configuration
type Config struct {
	Email      EmailConfig      `yaml:"email"`
	AI         AIConfig         `yaml:"ai"`
	Hugo       HugoConfig       `yaml:"hugo"`
	YouTube    YouTubeConfig    `yaml:"youtube"`
	BluesSky   BlueSkyConfig    `yaml:"bluesky"`
	Slack      SlackConfig      `yaml:"slack"`
	LinkedIn   LinkedInConfig   `yaml:"linkedin"`
	HackerNews HackerNewsConfig `yaml:"hackernews"`
	Defaults   DefaultsConfig   `yaml:"defaults"`
}

// EmailConfig represents email settings
type EmailConfig struct {
	From        string `yaml:"from"`
	ThumbnailTo string `yaml:"thumbnailTo"`
	EditTo      string `yaml:"editTo"`
	FinanceTo   string `yaml:"financeTo"`
}

// AIConfig represents AI settings
type AIConfig struct {
	Endpoint   string `yaml:"endpoint"`
	Deployment string `yaml:"deployment"`
	Model      string `yaml:"model"`
	APIKey     string `yaml:"apiKey"`
}

// HugoConfig represents Hugo settings
type HugoConfig struct {
	Path string `yaml:"path"`
	URL  string `yaml:"url"`
}

// YouTubeConfig represents YouTube settings
type YouTubeConfig struct {
	APIKey       string `yaml:"apiKey"`
	ChannelID    string `yaml:"channelId"`
	ClientID     string `yaml:"clientId"`
	ClientSecret string `yaml:"clientSecret"`
}

// BlueSkyConfig represents BlueSky settings
type BlueSkyConfig struct {
	Identifier string `yaml:"identifier"`
	URL        string `yaml:"url"`
	Password   string `yaml:"password"`
}

// SlackConfig represents Slack settings
type SlackConfig struct {
	Webhook string `yaml:"webhook"`
	Channel string `yaml:"channel"`
}

// LinkedInConfig represents LinkedIn settings
type LinkedInConfig struct {
	ClientID     string `yaml:"clientId"`
	ClientSecret string `yaml:"clientSecret"`
}

// HackerNewsConfig represents HackerNews settings
type HackerNewsConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// DefaultsConfig represents default settings
type DefaultsConfig struct {
	Category string `yaml:"category"`
	Tags     string `yaml:"tags"`
	Location string `yaml:"location"`
}

// LoadTestVideo loads a video fixture from the testdata directory
func LoadTestVideo(t testing.TB, filename string) Video {
	t.Helper()

	// Get the absolute path to the fixture
	path := filepath.Join("testdata", "videos", filename)
	absPath, err := filepath.Abs(path)
	if err != nil {
		// Try relative to the package
		pkgPath := filepath.Join("pkg", "testutil", "testdata", "videos", filename)
		absPath, err = filepath.Abs(pkgPath)
		if err != nil {
			t.Fatalf("Failed to resolve path to test video fixture: %v", err)
		}
	}

	// Read and parse the YAML file
	data, err := os.ReadFile(absPath)
	if err != nil {
		t.Fatalf("Failed to read test video fixture %s: %v", filename, err)
	}

	var video Video
	err = yaml.Unmarshal(data, &video)
	if err != nil {
		t.Fatalf("Failed to parse test video fixture %s: %v", filename, err)
	}

	return video
}

// LoadTestVideoRaw loads a video fixture as raw bytes
func LoadTestVideoRaw(t testing.TB, filename string) []byte {
	t.Helper()

	// Get the absolute path to the fixture
	path := filepath.Join("testdata", "videos", filename)
	absPath, err := filepath.Abs(path)
	if err != nil {
		// Try relative to the package
		pkgPath := filepath.Join("pkg", "testutil", "testdata", "videos", filename)
		absPath, err = filepath.Abs(pkgPath)
		if err != nil {
			t.Fatalf("Failed to resolve path to test video fixture: %v", err)
		}
	}

	// Read the YAML file
	data, err := os.ReadFile(absPath)
	if err != nil {
		t.Fatalf("Failed to read test video fixture %s: %v", filename, err)
	}

	return data
}

// LoadTestIndex loads an index fixture from the testdata directory
func LoadTestIndex(t testing.TB, filename string) []VideoIndex {
	t.Helper()

	// Get the absolute path to the fixture
	path := filepath.Join("testdata", "indexes", filename)
	absPath, err := filepath.Abs(path)
	if err != nil {
		// Try relative to the package
		pkgPath := filepath.Join("pkg", "testutil", "testdata", "indexes", filename)
		absPath, err = filepath.Abs(pkgPath)
		if err != nil {
			t.Fatalf("Failed to resolve path to test index fixture: %v", err)
		}
	}

	// Read and parse the YAML file
	data, err := os.ReadFile(absPath)
	if err != nil {
		t.Fatalf("Failed to read test index fixture %s: %v", filename, err)
	}

	var index []VideoIndex
	err = yaml.Unmarshal(data, &index)
	if err != nil {
		t.Fatalf("Failed to parse test index fixture %s: %v", filename, err)
	}

	return index
}

// LoadTestIndexRaw loads an index fixture as raw bytes
func LoadTestIndexRaw(t testing.TB, filename string) []byte {
	t.Helper()

	// Get the absolute path to the fixture
	path := filepath.Join("testdata", "indexes", filename)
	absPath, err := filepath.Abs(path)
	if err != nil {
		// Try relative to the package
		pkgPath := filepath.Join("pkg", "testutil", "testdata", "indexes", filename)
		absPath, err = filepath.Abs(pkgPath)
		if err != nil {
			t.Fatalf("Failed to resolve path to test index fixture: %v", err)
		}
	}

	// Read the YAML file
	data, err := os.ReadFile(absPath)
	if err != nil {
		t.Fatalf("Failed to read test index fixture %s: %v", filename, err)
	}

	return data
}

// LoadTestConfig loads a config fixture from the testdata directory
func LoadTestConfig(t testing.TB, filename string) Config {
	t.Helper()

	// Get the absolute path to the fixture
	path := filepath.Join("testdata", "configs", filename)
	absPath, err := filepath.Abs(path)
	if err != nil {
		// Try relative to the package
		pkgPath := filepath.Join("pkg", "testutil", "testdata", "configs", filename)
		absPath, err = filepath.Abs(pkgPath)
		if err != nil {
			t.Fatalf("Failed to resolve path to test config fixture: %v", err)
		}
	}

	// Read and parse the YAML file
	data, err := os.ReadFile(absPath)
	if err != nil {
		t.Fatalf("Failed to read test config fixture %s: %v", filename, err)
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		t.Fatalf("Failed to parse test config fixture %s: %v", filename, err)
	}

	return config
}

// LoadTestConfigRaw loads a config fixture as raw bytes
func LoadTestConfigRaw(t testing.TB, filename string) []byte {
	t.Helper()

	// Get the absolute path to the fixture
	path := filepath.Join("testdata", "configs", filename)
	absPath, err := filepath.Abs(path)
	if err != nil {
		// Try relative to the package
		pkgPath := filepath.Join("pkg", "testutil", "testdata", "configs", filename)
		absPath, err = filepath.Abs(pkgPath)
		if err != nil {
			t.Fatalf("Failed to resolve path to test config fixture: %v", err)
		}
	}

	// Read the YAML file
	data, err := os.ReadFile(absPath)
	if err != nil {
		t.Fatalf("Failed to read test config fixture %s: %v", filename, err)
	}

	return data
}

// ---- Video Phase Testing Helpers ----

// VideoPhaseConstants represent the possible phases of a video
type VideoPhaseConstants struct {
	PhaseIdeas            int
	PhaseStarted          int
	PhaseMaterialDone     int
	PhaseEditRequested    int
	PhasePublishPending   int
	PhasePublished        int
	PhaseDelayed          int
	PhaseSponsoredBlocked int
}

// DetermineVideoPhase calculates the phase of a video based on its workflow state
// This is a wrapper around video.CalculateVideoPhase for testing convenience
func DetermineVideoPhase(videoData storage.Video) int {
	return video.CalculateVideoPhase(videoData)
}
