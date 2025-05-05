# Testing Examples

This directory contains example code for different types of tests to help developers understand the testing patterns used in the YouTube Automation project.

## Example Files

- **unit_test_example.go** - Demonstrates basic unit testing techniques including table-driven tests and assertions.
- **mock_example.go** - Shows how to create and use mocks for external services like the YouTube API.
- **integration_test_example.go** - Illustrates integration testing of components working together.

## Using These Examples

These examples are for reference only and are not part of the actual application code. They demonstrate common testing patterns and how to structure different types of tests.

### Key Patterns Demonstrated

1. **Arrange-Act-Assert Pattern**
   - Set up test data and expectations
   - Call the function/method being tested
   - Verify the results match expectations

2. **Table-Driven Tests**
   - Multiple test cases in a single test function
   - Descriptive names for each test case
   - Reusable test structure

3. **Interface-Based Mocking**
   - Define interfaces for external dependencies
   - Create mock implementations for testing
   - Configure mock responses for different test scenarios
   - Track mock calls to verify behavior

4. **Testing Both Success and Error Paths**
   - Test the "happy path" (successful execution)
   - Test error conditions and edge cases
   - Verify correct error handling

5. **Test Fixtures**
   - Creating temporary test files
   - Setting up test environments
   - Cleaning up after tests

## Running the Examples

While these examples are provided for reference, they can be executed to see how they work:

```bash
# Run all examples
go test ./docs/examples

# Run a specific example
go test ./docs/examples -run TestFormatVideoTitle

# Run with verbose output
go test -v ./docs/examples
```

Note: You may need to create a Go module in the examples directory to run these tests directly.

## Related Documentation

See the main [testing.md](../testing.md) file for complete documentation on the project's testing approach. 