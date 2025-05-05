# Testing Guide for YouTube Automation

This document provides a comprehensive guide to the testing approach used in the YouTube Automation project, including how to run tests, write new tests, and work with the mock implementations.

## Table of Contents

- [Testing Philosophy and Approach](#testing-philosophy-and-approach)
- [Test Directory Structure](#test-directory-structure)
- [Running Tests](#running-tests)
- [Writing New Tests](#writing-new-tests)
- [Mock Implementations](#mock-implementations)
- [Test Fixtures](#test-fixtures)
- [Common Testing Patterns](#common-testing-patterns)
- [Test Coverage](#test-coverage)
- [Test Style Guide](#test-style-guide)
- [Brittle Tests and How to Avoid Them](#brittle-tests-and-how-to-avoid-them)
- [Test Maintenance Tools](#test-maintenance-tools)

## Testing Philosophy and Approach

The YouTube Automation project follows these testing principles:

1. **Unit Tests** - Each function or component is tested in isolation with appropriate mocks
2. **Integration Tests** - Testing interactions between components
3. **Mock-Based Testing** - External services like YouTube API are mocked to prevent actual API calls
4. **Test Coverage** - Aim for at least 80% code coverage while focusing on critical paths
5. **Test Readability** - Tests serve as documentation and should be clear and maintainable

## Test Directory Structure

Tests are organized following Go conventions:

- Test files are located in the same directory as the code they test
- Test file names follow the pattern `*_test.go` 
- Mock implementations are in the `pkg/mocks` directory
- Test utilities are in the `pkg/testutil` directory

## Running Tests

### Running All Tests

```bash
go test ./...
```

### Running Tests with Coverage

```bash
./scripts/coverage.sh
```

This script will:
- Run all tests with coverage tracking
- Generate a detailed coverage report by function
- Create an HTML report for visual inspection
- Compare the coverage against our 80% threshold
- Identify packages with the lowest coverage for improvement

### Running Specific Tests

```bash
# Run tests in a specific package
go test ./pkg/bluesky

# Run a specific test function
go test -run TestGetAdditionalInfo

# Run specific tests with verbose output
go test -v -run TestGetAdditionalInfo
```

## Writing New Tests

### Basic Test Structure

Follow this pattern when creating new tests:

```go
func TestMyFunction(t *testing.T) {
    // Arrange - Set up test data and expectations
    input := "test input"
    expected := "expected output"
    
    // Act - Call the function being tested
    result := MyFunction(input)
    
    // Assert - Verify the results
    if result != expected {
        t.Errorf("Expected %s but got %s", expected, result)
    }
}
```

### Table-Driven Tests

For functions that need multiple test cases, use table-driven tests:

```go
func TestMyFunction(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        {
            name:     "Valid input",
            input:    "valid",
            expected: "processed valid",
            wantErr:  false,
        },
        {
            name:     "Empty input",
            input:    "",
            expected: "",
            wantErr:  true,
        },
        // Add more test cases as needed
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := MyFunction(tt.input)
            
            if (err != nil) != tt.wantErr {
                t.Errorf("MyFunction() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            
            if result != tt.expected {
                t.Errorf("MyFunction() = %v, want %v", result, tt.expected)
            }
        })
    }
}
```

### Using Assertions

The project uses `github.com/stretchr/testify/assert` for cleaner assertions:

```go
import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestMyFunction(t *testing.T) {
    // Arrange
    input := "test"
    expected := "result"
    
    // Act
    result := MyFunction(input)
    
    // Assert
    assert.Equal(t, expected, result, "The result should match the expected value")
}
```

## Mock Implementations

### YouTube API Mocks

The YouTube API interactions are mocked to avoid making actual API calls during tests.

#### Example: Using the YouTube Service Mock

```go
func TestUploadVideo(t *testing.T) {
    // Setup mock service
    mockService := &mockYouTubeService{
        uploadResponse: "video123",
    }
    
    // Test with the mock
    videoID, err := mockService.uploadVideo(
        "Test Title",
        "Test Description",
        []string{"test", "video"},
        "22", // Entertainment category
        "video.mp4",
        "thumbnail.jpg",
    )
    
    // Assertions
    assert.NoError(t, err)
    assert.Equal(t, "video123", videoID)
}
```

### Email Mocks

#### Example: Using the Email Client Mock

```go
func TestSendEmail(t *testing.T) {
    // Setup mock email client
    mockClient := &MockEmailClient{
        SendResponse: nil, // No error
    }
    
    // Test with the mock
    err := SendEmailNotification(mockClient, "recipient@example.com", "Test Subject", "Test Body")
    
    // Assertions
    assert.NoError(t, err)
    assert.Equal(t, 1, mockClient.SendCallCount, "Send should be called once")
    assert.Equal(t, "recipient@example.com", mockClient.LastRecipient)
}
```

### File System Mocks

#### Example: Using the MockFileSystem

```go
func TestReadVideoConfig(t *testing.T) {
    // Setup mock filesystem
    mockFS := &mocks.MockFileSystem{
        Files: map[string][]byte{
            "video.yaml": []byte(`title: Test Video`),
        },
        Errors: map[string]error{
            "missing.yaml": os.ErrNotExist,
        },
    }
    
    // Use in tests
    data, err := mockFS.ReadFile("video.yaml")
    
    // Assertions
    assert.NoError(t, err)
    assert.Contains(t, string(data), "Test Video")
}
```

## Test Fixtures

Test fixtures are located in appropriate directories and contain test data used in tests.

### YAML Test Data

Example YAML fixtures for testing video upload configuration:

```yaml
# Example from pkg/testutil/testdata/video.yaml
title: Test Video
description: A test video description
tags:
  - test
  - sample
category: 22  # Entertainment
videoPath: ./test-video.mp4
thumbnailPath: ./thumbnail.jpg
```

### Using Test Fixtures

```go
func TestParseVideoConfig(t *testing.T) {
    data, err := os.ReadFile("../testutil/testdata/video.yaml")
    assert.NoError(t, err)
    
    config, err := ParseVideoConfig(data)
    assert.NoError(t, err)
    assert.Equal(t, "Test Video", config.Title)
}
```

## Common Testing Patterns

### Testing Error Cases

Ensure you test both success and error paths:

```go
func TestErrorHandling(t *testing.T) {
    // Test with valid input
    result, err := ProcessInput("valid")
    assert.NoError(t, err)
    assert.NotEmpty(t, result)
    
    // Test with invalid input
    result, err = ProcessInput("")
    assert.Error(t, err)
    assert.Empty(t, result)
    assert.Contains(t, err.Error(), "input cannot be empty")
}
```

### Testing with Environment Variables

When testing code that uses environment variables:

```go
func TestWithEnvVariables(t *testing.T) {
    // Save original env and restore after test
    originalValue := os.Getenv("API_KEY")
    defer os.Setenv("API_KEY", originalValue)
    
    // Set test environment
    os.Setenv("API_KEY", "test-key")
    
    // Run test with the modified environment
    result := GetAPIKey()
    assert.Equal(t, "test-key", result)
}
```

### Testing CLI Arguments

```go
func TestFlagParsing(t *testing.T) {
    // Save original args and restore after test
    originalArgs := os.Args
    defer func() { os.Args = originalArgs }()
    
    // Set test arguments
    os.Args = []string{"cmd", "--flag", "value"}
    
    // Test parsing
    flags := ParseFlags()
    assert.Equal(t, "value", flags.Get("flag"))
}
```

## Test Coverage

The project aims for a test coverage goal of 80%. To check current test coverage:

```bash
./scripts/coverage.sh
```

This will:
1. Run all tests with coverage tracking
2. Generate a detailed coverage report by function
3. Create an HTML report for visual inspection

### Improving Test Coverage

When adding new features or fixing bugs:

1. Write tests before or alongside the implementation
2. Ensure you cover both the "happy path" and error conditions
3. Use the coverage report to identify untested areas
4. Focus on critical paths and complex logic first

### Test Coverage by Module

Current test coverage by module (as of last update):

- Main package: 25.6%
- pkg/bluesky: 89.2%
- pkg/mocks: 79.2%
- pkg/slack: 37.5%
- pkg/testutil: 76.8%

Overall coverage: 36.3%

### Coverage Goals

Each new PR should maintain or improve the current code coverage. Focus areas for improvement:

1. Main package functions
2. Error handling paths
3. Edge cases in data processing functions

## Test Style Guide

### Naming Conventions
- Test functions should be named `Test<FunctionName>` or `Test<Behavior>`
- Test cases in table-driven tests should have descriptive names
- Helper functions should be prefixed with `test` or `assert`

### Structure
- Use table-driven tests for multiple test cases
- Keep test files in the same package as the code they test
- Use subtests for related test cases
- Keep test setup and teardown code separate from assertions

### Assertions
- Use clear, specific assertions
- Include helpful error messages in assertions
- Prefer specific assertions (e.g., `assert.Equal`) over generic ones

### Test Data
- Use meaningful test data that represents real-world scenarios
- Avoid hardcoded values without clear explanation
- Define test data at the top of the test or in separate test fixtures
- Document the purpose of edge-case test data

### Mocking
- Create mocks that accurately reflect the behavior of the real implementation
- Document expected behavior of mocks
- Keep mock implementations simple and focused
- Use interface-based mocking when possible

### Test Performance
- Tests should be fast and not dependent on external services
- Avoid unnecessary computation in tests
- Be mindful of test setup and teardown costs
- Use parallel tests when appropriate with t.Parallel()

## Brittle Tests and How to Avoid Them

Brittle tests are tests that fail due to changes unrelated to the functionality being tested. Here are guidelines to avoid brittle tests:

1. **Avoid Implementation Details**: Test behavior, not implementation details that might change

2. **Flexible Assertions**: Use substring matches or regexp instead of exact string matches when appropriate

3. **Avoid Time Dependencies**: Be cautious with time-dependent tests and use time mocking when needed

4. **Resilient Mocks**: Create mocks that don't break when the underlying implementation changes in non-essential ways

5. **Focused Tests**: Each test should test one specific behavior

6. **Independent Tests**: Tests should not depend on other tests or running order

The `scripts/find_brittle_tests.sh` script can help identify potentially brittle tests in the codebase.

## Test Maintenance Tools

This project includes several tools to help maintain test quality and consistency:

### Linting Configuration for Tests

We use a dedicated linting configuration for tests in `.golangci-test.yml`. This configuration includes linters specifically chosen for test files:

```bash
# Run linters on test files
golangci-lint run --config=.golangci-test.yml
```

The linting configuration focuses on:
- Code formatting and import organization
- Basic error checking
- Go best practices
- Test package structure

### Finding Brittle Tests

The `scripts/find_brittle_tests.sh` script helps identify potentially brittle tests in the codebase:

```bash
./scripts/find_brittle_tests.sh
```

This script looks for common patterns that might indicate brittle tests:
1. Hard-coded string comparisons
2. Time-dependent tests
3. Magic numbers
4. Excessive test setup
5. Environment dependencies
6. Tests that may depend on execution order

The script generates a report in `brittle_tests_report.txt` that can guide test refactoring efforts.

### Test Coverage Reporting

The `scripts/coverage.sh` script provides detailed test coverage information:

```bash
./scripts/coverage.sh
```

Use this script to identify areas of the codebase that need additional test coverage.

---

This documentation is maintained as part of Task #11 in the project's Taskmaster tasks list. For questions or suggestions about this testing documentation, please open an issue or contact the maintainers. 