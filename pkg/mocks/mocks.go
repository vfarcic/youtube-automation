// Package mocks provides mockable implementations of external dependencies
// used throughout the YouTube Automation project to facilitate testing.
package mocks

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ---- File System Mocks ----

// MockFileSystem provides a simple in-memory filesystem implementation
// for testing file operations without using the actual filesystem.
type MockFileSystem struct {
	Files     map[string][]byte
	Dirs      map[string]bool
	FileInfos map[string]MockFileInfo
	Errors    map[string]error
}

// NewMockFileSystem creates a new MockFileSystem instance
func NewMockFileSystem() *MockFileSystem {
	return &MockFileSystem{
		Files:     make(map[string][]byte),
		Dirs:      make(map[string]bool),
		FileInfos: make(map[string]MockFileInfo),
		Errors:    make(map[string]error),
	}
}

// ReadFile mocks reading a file from the filesystem
func (m *MockFileSystem) ReadFile(path string) ([]byte, error) {
	if err, exists := m.Errors[path]; exists && err != nil {
		return nil, err
	}

	data, exists := m.Files[path]
	if !exists {
		return nil, os.ErrNotExist
	}
	return data, nil
}

// WriteFile mocks writing data to a file
func (m *MockFileSystem) WriteFile(path string, data []byte, perm os.FileMode) error {
	if err, exists := m.Errors[path]; exists && err != nil {
		return err
	}

	// Ensure parent directory exists
	dir := filepath.Dir(path)
	m.Dirs[dir] = true

	// Store the file
	m.Files[path] = data

	// Create default file info if it doesn't exist
	if _, exists := m.FileInfos[path]; !exists {
		m.FileInfos[path] = MockFileInfo{
			FileName:     filepath.Base(path),
			IsDirectory:  false,
			FileSize:     int64(len(data)),
			FileMode:     perm,
			ModifiedTime: time.Now(),
		}
	}

	return nil
}

// MkdirAll mocks creating a directory and its parents
func (m *MockFileSystem) MkdirAll(path string, perm os.FileMode) error {
	if err, exists := m.Errors[path]; exists && err != nil {
		return err
	}

	// Create the directory and all parent directories
	path = filepath.Clean(path)
	m.Dirs[path] = true

	// Create parent directories
	dir := path
	for dir != "." && dir != "/" {
		dir = filepath.Dir(dir)
		if dir != "." && dir != "/" {
			m.Dirs[dir] = true
		}
	}

	// Create directory file info
	m.FileInfos[path] = MockFileInfo{
		FileName:     filepath.Base(path),
		IsDirectory:  true,
		FileSize:     0,
		FileMode:     perm,
		ModifiedTime: time.Now(),
	}

	return nil
}

// RemoveAll mocks removing a file or directory
func (m *MockFileSystem) RemoveAll(path string) error {
	if err, exists := m.Errors[path]; exists && err != nil {
		return err
	}

	// Check if path exists
	_, existsAsFile := m.Files[path]
	_, existsAsDir := m.Dirs[path]

	if !existsAsFile && !existsAsDir {
		return os.ErrNotExist
	}

	// Remove the file if it exists
	if existsAsFile {
		delete(m.Files, path)
		delete(m.FileInfos, path)
	}

	// Remove the directory and any children
	if existsAsDir {
		prefix := path + string(os.PathSeparator)

		// Remove child files
		for filePath := range m.Files {
			if filePath == path || strings.HasPrefix(filePath, prefix) {
				delete(m.Files, filePath)
				delete(m.FileInfos, filePath)
			}
		}

		// Remove child directories
		for dirPath := range m.Dirs {
			if dirPath == path || strings.HasPrefix(dirPath, prefix) {
				delete(m.Dirs, dirPath)
				delete(m.FileInfos, dirPath)
			}
		}
	}

	return nil
}

// Stat mocks the os.Stat function, returning file info
func (m *MockFileSystem) Stat(path string) (os.FileInfo, error) {
	if err, exists := m.Errors[path]; exists && err != nil {
		return nil, err
	}

	// Check if the file exists
	_, fileExists := m.Files[path]
	_, dirExists := m.Dirs[path]

	if !fileExists && !dirExists {
		return nil, os.ErrNotExist
	}

	// Return the file info
	if info, exists := m.FileInfos[path]; exists {
		return &info, nil
	}

	// If we have a file but no explicit file info, create a basic one
	if fileExists {
		info := MockFileInfo{
			FileName:     filepath.Base(path),
			IsDirectory:  false,
			FileSize:     int64(len(m.Files[path])),
			FileMode:     0644,
			ModifiedTime: time.Now(),
		}
		m.FileInfos[path] = info
		return &info, nil
	}

	// If we have a directory but no explicit file info, create a basic one
	if dirExists {
		info := MockFileInfo{
			FileName:     filepath.Base(path),
			IsDirectory:  true,
			FileSize:     0,
			FileMode:     0755,
			ModifiedTime: time.Now(),
		}
		m.FileInfos[path] = info
		return &info, nil
	}

	return nil, os.ErrNotExist
}

// MockFileInfo implements os.FileInfo for testing
type MockFileInfo struct {
	FileName     string
	IsDirectory  bool
	FileSize     int64
	FileMode     os.FileMode
	ModifiedTime time.Time
}

// Name returns the base name of the file
func (m *MockFileInfo) Name() string {
	return m.FileName
}

// Size returns the size of the file in bytes
func (m *MockFileInfo) Size() int64 {
	return m.FileSize
}

// Mode returns the file mode bits
func (m *MockFileInfo) Mode() os.FileMode {
	return m.FileMode
}

// ModTime returns the modification time
func (m *MockFileInfo) ModTime() time.Time {
	return m.ModifiedTime
}

// IsDir returns whether the file is a directory
func (m *MockFileInfo) IsDir() bool {
	return m.IsDirectory
}

// Sys returns system-dependent information about the file
func (m *MockFileInfo) Sys() interface{} {
	return nil
}

// ---- HTTP Mocks ----

// MockHTTPClient is a mockable HTTP client for testing
type MockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

// Do executes the mock HTTP request
func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	if m.DoFunc != nil {
		return m.DoFunc(req)
	}
	return nil, errors.New("mock HTTP client DoFunc not implemented")
}

// NewMockHTTPClient creates a new MockHTTPClient with the given response
func NewMockHTTPClient(response *http.Response, err error) *MockHTTPClient {
	return &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return response, err
		},
	}
}

// MockResponse creates a mock HTTP response with the given status code and body
func MockResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
}

// MockJSONResponse creates a mock HTTP response with JSON content
func MockJSONResponse(statusCode int, data interface{}, err error) (*http.Response, error) {
	if err != nil {
		return nil, err
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	response := MockResponse(statusCode, string(jsonData))
	response.Header.Set("Content-Type", "application/json")
	return response, nil
}

// NewMockHTTPServer creates a test server with route handlers
func NewMockHTTPServer(handlers map[string]http.HandlerFunc) *httptest.Server {
	mux := http.NewServeMux()

	for route, handler := range handlers {
		mux.HandleFunc(route, handler)
	}

	return httptest.NewServer(mux)
}

// ---- Command Execution Mocks ----

// MockCommandResult represents the result of a command execution
type MockCommandResult struct {
	Output string
	Error  error
}

// MockCommandExecutor provides a mock implementation for executing commands
type MockCommandExecutor struct {
	ExpectedCommands map[string]MockCommandResult
	ExecutedCommands []string
}

// NewMockCommandExecutor creates a new MockCommandExecutor
func NewMockCommandExecutor() *MockCommandExecutor {
	return &MockCommandExecutor{
		ExpectedCommands: make(map[string]MockCommandResult),
		ExecutedCommands: []string{},
	}
}

// AddCommand adds an expected command with its result
func (m *MockCommandExecutor) AddCommand(cmd string, output string, err error) {
	m.ExpectedCommands[cmd] = MockCommandResult{
		Output: output,
		Error:  err,
	}
}

// Execute pretends to execute a command and returns predefined results
func (m *MockCommandExecutor) Execute(cmd string, args ...string) (string, error) {
	fullCmd := cmd
	if len(args) > 0 {
		fullCmd += " " + strings.Join(args, " ")
	}

	m.ExecutedCommands = append(m.ExecutedCommands, fullCmd)

	if result, exists := m.ExpectedCommands[fullCmd]; exists {
		return result.Output, result.Error
	}

	return "", errors.New("unexpected command: " + fullCmd)
}

// ---- User I/O Mocks ----

// MockIO provides a mock for user input/output operations
type MockIO struct {
	InputResponses []string
	InputIndex     int
	OutputCapture  []string
}

// NewMockIO creates a new MockIO instance
func NewMockIO(responses []string) *MockIO {
	return &MockIO{
		InputResponses: responses,
		InputIndex:     0,
		OutputCapture:  []string{},
	}
}

// ReadInput returns the next predefined input response
func (m *MockIO) ReadInput() string {
	if m.InputIndex >= len(m.InputResponses) {
		return ""
	}

	response := m.InputResponses[m.InputIndex]
	m.InputIndex++
	return response
}

// WriteOutput captures output for later inspection
func (m *MockIO) WriteOutput(s string) {
	m.OutputCapture = append(m.OutputCapture, s)
}

// GetCapturedOutput returns all captured output
func (m *MockIO) GetCapturedOutput() []string {
	return m.OutputCapture
}

// GetCapturedOutputString returns all captured output as a single string
func (m *MockIO) GetCapturedOutputString() string {
	return strings.Join(m.OutputCapture, "\n")
}

// ClearCapturedOutput clears the captured output
func (m *MockIO) ClearCapturedOutput() {
	m.OutputCapture = []string{}
}
