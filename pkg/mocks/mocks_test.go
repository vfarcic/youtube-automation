package mocks

import (
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
)

func TestMockFileSystem(t *testing.T) {
	fs := NewMockFileSystem()

	// Test WriteFile and ReadFile
	testPath := "/test/file.txt"
	testContent := []byte("test content")

	err := fs.WriteFile(testPath, testContent, 0644)
	if err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}

	data, err := fs.ReadFile(testPath)
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}

	if string(data) != string(testContent) {
		t.Errorf("Expected content %q, got %q", testContent, data)
	}

	// Test Stat
	info, err := fs.Stat(testPath)
	if err != nil {
		t.Fatalf("Stat error: %v", err)
	}

	if info.Name() != "file.txt" {
		t.Errorf("Expected file name 'file.txt', got %q", info.Name())
	}

	if info.Size() != int64(len(testContent)) {
		t.Errorf("Expected size %d, got %d", len(testContent), info.Size())
	}

	if info.IsDir() {
		t.Errorf("Expected file not to be a directory")
	}

	// Test error cases
	fs.Errors["/error/path"] = errors.New("simulated error")

	_, err = fs.ReadFile("/error/path")
	if err == nil || err.Error() != "simulated error" {
		t.Errorf("Expected simulated error, got %v", err)
	}

	_, err = fs.ReadFile("/nonexistent")
	if err != os.ErrNotExist {
		t.Errorf("Expected ErrNotExist, got %v", err)
	}

	// Test MkdirAll
	err = fs.MkdirAll("/test/dir", 0755)
	if err != nil {
		t.Fatalf("MkdirAll error: %v", err)
	}

	info, err = fs.Stat("/test/dir")
	if err != nil {
		t.Fatalf("Stat error after MkdirAll: %v", err)
	}

	if !info.IsDir() {
		t.Errorf("Expected path to be a directory")
	}

	// Test RemoveAll
	err = fs.RemoveAll("/test")
	if err != nil {
		t.Fatalf("RemoveAll error: %v", err)
	}

	_, err = fs.Stat("/test/file.txt")
	if err != os.ErrNotExist {
		t.Errorf("Expected file to be removed, got error: %v", err)
	}

	_, err = fs.Stat("/test/dir")
	if err != os.ErrNotExist {
		t.Errorf("Expected directory to be removed, got error: %v", err)
	}
}

func TestMockHTTPClient(t *testing.T) {
	expectedBody := "test response"
	mockResponse := MockResponse(200, expectedBody)
	client := NewMockHTTPClient(mockResponse, nil)

	req, _ := http.NewRequest("GET", "http://example.com", nil)
	resp, err := client.Do(req)

	if err != nil {
		t.Fatalf("HTTP request error: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	if string(body) != expectedBody {
		t.Errorf("Expected body %q, got %q", expectedBody, body)
	}

	// Test JSON response
	type TestData struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	testData := TestData{Name: "test", Value: 42}
	jsonResp, err := MockJSONResponse(201, testData, nil)

	if err != nil {
		t.Fatalf("JSON response error: %v", err)
	}

	if jsonResp.StatusCode != 201 {
		t.Errorf("Expected JSON status code 201, got %d", jsonResp.StatusCode)
	}

	if jsonResp.Header.Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type header to be application/json")
	}

	// Test error case
	expectedError := errors.New("HTTP error")
	errorClient := NewMockHTTPClient(nil, expectedError)

	_, err = errorClient.Do(req)
	if err != expectedError {
		t.Errorf("Expected HTTP error, got %v", err)
	}

	// Test HTTP server
	handlers := map[string]http.HandlerFunc{
		"/test": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			w.Write([]byte("test server response"))
		},
	}

	server := NewMockHTTPServer(handlers)
	defer server.Close()

	resp, err = http.Get(server.URL + "/test")
	if err != nil {
		t.Fatalf("Server test error: %v", err)
	}

	serverBody, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	if string(serverBody) != "test server response" {
		t.Errorf("Expected server response %q, got %q", "test server response", serverBody)
	}
}

func TestMockCommandExecutor(t *testing.T) {
	executor := NewMockCommandExecutor()

	// Add expected commands
	executor.AddCommand("ls", "file1.txt\nfile2.txt", nil)
	executor.AddCommand("git status", "On branch main", nil)
	executor.AddCommand("error command", "", errors.New("command failed"))

	// Test successful commands
	output, err := executor.Execute("ls")
	if err != nil {
		t.Errorf("Expected no error for 'ls', got %v", err)
	}
	if output != "file1.txt\nfile2.txt" {
		t.Errorf("Expected output %q, got %q", "file1.txt\nfile2.txt", output)
	}

	output, err = executor.Execute("git", "status")
	if err != nil {
		t.Errorf("Expected no error for 'git status', got %v", err)
	}
	if output != "On branch main" {
		t.Errorf("Expected output %q, got %q", "On branch main", output)
	}

	// Test error command
	output, err = executor.Execute("error", "command")
	if err == nil || err.Error() != "command failed" {
		t.Errorf("Expected error 'command failed', got %v", err)
	}

	// Test unexpected command
	output, err = executor.Execute("unknown")
	if err == nil || !strings.Contains(err.Error(), "unexpected command") {
		t.Errorf("Expected 'unexpected command' error, got %v", err)
	}

	// Test executed commands tracking
	if len(executor.ExecutedCommands) != 4 {
		t.Errorf("Expected 4 executed commands, got %d", len(executor.ExecutedCommands))
	}

	expectedCommands := []string{"ls", "git status", "error command", "unknown"}
	for i, cmd := range expectedCommands {
		if i >= len(executor.ExecutedCommands) {
			t.Errorf("Missing executed command at index %d", i)
			continue
		}
		if executor.ExecutedCommands[i] != cmd {
			t.Errorf("Expected command %q at index %d, got %q", cmd, i, executor.ExecutedCommands[i])
		}
	}
}

func TestMockIO(t *testing.T) {
	responses := []string{"yes", "42", "quit"}
	mockIO := NewMockIO(responses)

	// Test input responses
	for i, expected := range responses {
		input := mockIO.ReadInput()
		if input != expected {
			t.Errorf("Expected input %q at index %d, got %q", expected, i, input)
		}
	}

	// Test empty response after all inputs are consumed
	input := mockIO.ReadInput()
	if input != "" {
		t.Errorf("Expected empty input after all responses consumed, got %q", input)
	}

	// Test output capture
	outputs := []string{
		"Starting process...",
		"Processing complete.",
		"Errors: 0",
	}

	for _, output := range outputs {
		mockIO.WriteOutput(output)
	}

	captured := mockIO.GetCapturedOutput()
	if len(captured) != len(outputs) {
		t.Errorf("Expected %d captured outputs, got %d", len(outputs), len(captured))
	}

	for i, expected := range outputs {
		if i >= len(captured) {
			t.Errorf("Missing captured output at index %d", i)
			continue
		}
		if captured[i] != expected {
			t.Errorf("Expected output %q at index %d, got %q", expected, i, captured[i])
		}
	}

	// Test captured output string
	expectedString := strings.Join(outputs, "\n")
	if mockIO.GetCapturedOutputString() != expectedString {
		t.Errorf("Expected captured string %q, got %q", expectedString, mockIO.GetCapturedOutputString())
	}

	// Test clear captured output
	mockIO.ClearCapturedOutput()
	if len(mockIO.GetCapturedOutput()) != 0 {
		t.Errorf("Expected empty captured output after clear, got %d items", len(mockIO.GetCapturedOutput()))
	}
}
