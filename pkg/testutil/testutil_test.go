package testutil

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSetupAndTeardownTestDir(t *testing.T) {
	// Test that a directory is created and then removed
	dir := SetupTestDir(t)

	// Verify directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Fatalf("Test directory was not created: %v", err)
	}

	// Teardown the directory
	TeardownTestDir(t, dir)

	// Verify directory no longer exists
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatalf("Test directory was not removed")
	}
}

func TestWriteTestFile(t *testing.T) {
	// Setup
	dir := SetupTestDir(t)
	defer TeardownTestDir(t, dir)

	// Test writing a file
	content := "test content"
	filename := "test.txt"
	path := WriteTestFile(t, dir, filename, content)

	// Verify file was written with correct content
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	if string(data) != content {
		t.Errorf("File content mismatch. Expected: %s, Got: %s", content, string(data))
	}
}

func TestCopyFile(t *testing.T) {
	// Setup
	dir := SetupTestDir(t)
	defer TeardownTestDir(t, dir)

	// Create source file
	srcContent := "source file content"
	srcPath := WriteTestFile(t, dir, "source.txt", srcContent)

	// Create destination path
	dstPath := filepath.Join(dir, "destination.txt")

	// Test copying
	CopyFile(t, srcPath, dstPath)

	// Verify destination file
	data, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("Failed to read destination file: %v", err)
	}

	if string(data) != srcContent {
		t.Errorf("File content mismatch. Expected: %s, Got: %s", srcContent, string(data))
	}
}

func TestCreateYAMLFixture(t *testing.T) {
	// Setup
	dir := SetupTestDir(t)
	defer TeardownTestDir(t, dir)

	// Test creating YAML fixture
	yamlContent := "key: value\narray:\n  - item1\n  - item2"
	path := CreateYAMLFixture(t, dir, "test.yaml", yamlContent)

	// Verify file content
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read YAML fixture: %v", err)
	}

	if string(data) != yamlContent {
		t.Errorf("YAML content mismatch. Expected: %s, Got: %s", yamlContent, string(data))
	}
}

func TestCreateJSONFixture(t *testing.T) {
	// Setup
	dir := SetupTestDir(t)
	defer TeardownTestDir(t, dir)

	// Test data
	testData := map[string]interface{}{
		"name":   "Test Name",
		"values": []int{1, 2, 3},
		"nested": map[string]string{
			"key": "value",
		},
	}

	// Create JSON fixture
	path := CreateJSONFixture(t, dir, "test.json", testData)

	// Read and parse the fixture
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read JSON fixture: %v", err)
	}

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Failed to parse JSON fixture: %v", err)
	}

	// Verify the top-level fields
	AssertEqual(t, "Test Name", result["name"])

	// Values array requires type assertions to compare properly
	values, ok := result["values"].([]interface{})
	if !ok {
		t.Fatal("Failed to assert values as array")
	}

	AssertEqual(t, float64(1), values[0])
	AssertEqual(t, float64(2), values[1])
	AssertEqual(t, float64(3), values[2])

	// Nested map requires type assertions
	nested, ok := result["nested"].(map[string]interface{})
	if !ok {
		t.Fatal("Failed to assert nested as map")
	}

	AssertEqual(t, "value", nested["key"])
}

func TestMockHTTPServer(t *testing.T) {
	// Create a handler that returns a simple response
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	})

	// Create the mock server
	server := MockHTTPServer(t, handler)

	// Make a request to the server
	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("Failed to make request to mock server: %v", err)
	}
	defer resp.Body.Close()

	// Verify response
	AssertEqual(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	AssertEqual(t, "test response", string(body))
}

func TestMockHTTPResponse(t *testing.T) {
	// Create a mock response
	resp := MockHTTPResponse(t, http.StatusCreated, "test body")

	// Verify response properties
	AssertEqual(t, http.StatusCreated, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	AssertEqual(t, "test body", string(body))
}

func TestMockJSONResponse(t *testing.T) {
	// Create a mock JSON response
	data := map[string]interface{}{
		"status": "success",
		"data":   []string{"item1", "item2"},
	}

	resp := MockJSONResponse(t, http.StatusOK, data)

	// Verify response properties
	AssertEqual(t, http.StatusOK, resp.StatusCode)
	AssertEqual(t, "application/json", resp.Header.Get("Content-Type"))

	// Read and parse the body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	AssertEqual(t, "success", result["status"])

	// Check array content
	items, ok := result["data"].([]interface{})
	if !ok {
		t.Fatal("Failed to assert data as array")
	}

	AssertEqual(t, "item1", items[0])
	AssertEqual(t, "item2", items[1])
}

func TestCaptureOutput(t *testing.T) {
	// Test capturing stdout
	output := CaptureOutput(t, func() {
		fmt.Println("Captured output")
		fmt.Print("Another line")
	})

	if !strings.Contains(output, "Captured output") {
		t.Errorf("Output mismatch. Expected to contain: %q, Got: %q", "Captured output", output)
	}

	if !strings.Contains(output, "Another line") {
		t.Errorf("Output mismatch. Expected to contain: %q, Got: %q", "Another line", output)
	}
}

// simpleTestingT implements the Tester interface for testing assertion functions
type simpleTestingT struct {
	failed      bool
	messages    []string
	helperCalls int
}

func (s *simpleTestingT) Errorf(format string, args ...interface{}) {
	s.failed = true
	s.messages = append(s.messages, fmt.Sprintf(format, args...))
}

func (s *simpleTestingT) Fatalf(format string, args ...interface{}) {
	s.failed = true
	s.messages = append(s.messages, fmt.Sprintf(format, args...))
}

func (s *simpleTestingT) Helper() {
	s.helperCalls++
}

func (s *simpleTestingT) Fatal(args ...interface{}) {
	s.failed = true
	s.messages = append(s.messages, fmt.Sprint(args...))
}

func (s *simpleTestingT) Error(args ...interface{}) {
	s.failed = true
	s.messages = append(s.messages, fmt.Sprint(args...))
}

func (s *simpleTestingT) Logf(format string, args ...interface{}) {
	// Do nothing, just record
	s.messages = append(s.messages, fmt.Sprintf(format, args...))
}

func TestAssertEqual(t *testing.T) {
	// Create a mock testing.T to capture failure
	mockT := &simpleTestingT{}

	// Test matching values (should not fail)
	AssertEqual(mockT, 42, 42)
	if mockT.failed {
		t.Error("AssertEqual failed for equal values")
	}

	// Reset mock
	mockT = &simpleTestingT{}

	// Test non-matching values (should fail)
	AssertEqual(mockT, 42, 43)
	if !mockT.failed {
		t.Error("AssertEqual did not fail for unequal values")
	}
}

func TestAssertNotEqual(t *testing.T) {
	// Create a mock testing.T to capture failure
	mockT := &simpleTestingT{}

	// Test non-matching values (should not fail)
	AssertNotEqual(mockT, 42, 43)
	if mockT.failed {
		t.Error("AssertNotEqual failed for different values")
	}

	// Reset mock
	mockT = &simpleTestingT{}

	// Test matching values (should fail)
	AssertNotEqual(mockT, 42, 42)
	if !mockT.failed {
		t.Error("AssertNotEqual did not fail for equal values")
	}
}

func TestAssertTrue(t *testing.T) {
	// Create a mock testing.T to capture failure
	mockT := &simpleTestingT{}

	// Test true condition (should not fail)
	AssertTrue(mockT, true)
	if mockT.failed {
		t.Error("AssertTrue failed for true condition")
	}

	// Reset mock
	mockT = &simpleTestingT{}

	// Test false condition (should fail)
	AssertTrue(mockT, false)
	if !mockT.failed {
		t.Error("AssertTrue did not fail for false condition")
	}
}

func TestAssertFalse(t *testing.T) {
	// Create a mock testing.T to capture failure
	mockT := &simpleTestingT{}

	// Test false condition (should not fail)
	AssertFalse(mockT, false)
	if mockT.failed {
		t.Error("AssertFalse failed for false condition")
	}

	// Reset mock
	mockT = &simpleTestingT{}

	// Test true condition (should fail)
	AssertFalse(mockT, true)
	if !mockT.failed {
		t.Error("AssertFalse did not fail for true condition")
	}
}

func TestAssertContains(t *testing.T) {
	// Create a mock testing.T to capture failure
	mockT := &simpleTestingT{}

	// Test string contains substring (should not fail)
	AssertContains(mockT, "Hello, World!", "World")
	if mockT.failed {
		t.Error("AssertContains failed when substring was present")
	}

	// Reset mock
	mockT = &simpleTestingT{}

	// Test string does not contain substring (should fail)
	AssertContains(mockT, "Hello, World!", "Universe")
	if !mockT.failed {
		t.Error("AssertContains did not fail when substring was absent")
	}
}

func TestCompareYAML(t *testing.T) {
	// Create a mock testing.T to capture failure
	mockT := &simpleTestingT{}

	// Test identical YAML strings (should not fail)
	yaml1 := "key: value\nlist:\n  - item1\n  - item2"
	CompareYAML(mockT, yaml1, yaml1)
	if mockT.failed {
		t.Error("CompareYAML failed for identical YAML")
	}

	// Reset mock
	mockT = &simpleTestingT{}

	// Test different YAML strings (should fail)
	yaml2 := "key: different\nlist:\n  - item1\n  - item2"
	CompareYAML(mockT, yaml1, yaml2)
	if !mockT.failed {
		t.Error("CompareYAML did not fail for different YAML")
	}
}
