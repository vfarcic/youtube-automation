# [IMPLEMENTED] YouTube Automation Project - Unit Testing Requirements
# Status: Fully implemented as of May 2024
# Implementation details can be found in docs/testing.md and docs/test-coverage.md

<context>
# Overview  
This document outlines the requirements for implementing comprehensive unit tests for the YouTube Automation project. The goal is to ensure all code modules are thoroughly tested without making calls to external APIs.

# Core Features  
- Isolated unit tests for all Go packages and modules
- Test coverage for critical business logic
- Mock implementations for external dependencies
- Test data fixtures for consistent testing
- Test utilities for common test operations
</context>
<PRD>
# YouTube Automation Project - Unit Testing Requirements

## Project Overview
YouTube Automation is a CLI tool for managing the YouTube video creation process, from initial planning through publication and promotion across various platforms. The application needs comprehensive unit tests to ensure reliability and facilitate future development.

## Testing Objectives
- Achieve minimum 80% code coverage across all non-vendor code
- Ensure all business logic is thoroughly tested
- Validate input/output handling for all functions
- Test error conditions and edge cases
- Create isolated tests that do not rely on external services

## Technical Implementation Requirements

### 1. Test Infrastructure & Utilities

#### 1.1 Mock Framework
- Implement a standardized approach for mocking external dependencies
- Create reusable mock implementations ONLY for external API calls and services:
  - HTTP requests/responses to third-party services
  - Command executions to external programs
  - User input/output
- Use Go's `httptest` package for HTTP interaction tests
- Utilize table-driven tests for comprehensive test coverage
- Do NOT mock file system operations - use real file system with temporary directories instead

#### 1.2 Test Helpers
- Create utility functions for:
  - Setting up and tearing down test environments with temporary directories 
  - Generating test fixtures and sample data
  - Validating test outputs
  - Comparing complex structures
- Organize helpers in a dedicated `testutil` package
- Include helpers for creating and cleaning up temporary test directories

#### 1.3 Test Data
- Create YAML fixtures for:
  - Sample video definitions
  - Settings configurations
  - Index data structures
- Store test data in a `testdata` directory
- Include both valid and invalid test cases

### 2. Core Module Testing

#### 2.1 YAML Operations Testing
- Test `GetVideo`, `WriteVideo`, `GetIndex`, and `WriteIndex` functions
- Verify correct parsing of YAML structures
- Test handling of malformed YAML data
- Validate error handling for file operations
- Test edge cases (empty files, permission issues)
- Use real file system with temporary test directories

#### 2.2 CLI Argument Handling
- Test command-line flag parsing and validation
- Verify environment variable handling
- Test required flag validation
- Validate settings merging from files and flags

#### 2.3 Choices Module Testing
- Test the interactive menu functionality
- Mock user input for testing UI flows
- Verify state transitions between menus
- Test validation of user inputs
- Verify color formatting and display logic
- Use real file system operations for any file interactions

### 3. Feature-Specific Testing

#### 3.1 Video Management Testing
- Test video creation, editing, and deletion using real file system operations
- Verify phase transitions (Init → Work → Define → Edit → Publish)
- Test task completion tracking
- Validate video metadata handling
- Test filtering and sorting functionality

#### 3.2 External Service Integration Testing
- Create mock implementations ONLY for external service APIs, with dedicated tasks for each service:
  - YouTube API interactions
  - Email sending functionality
  - Hugo site integration
  - Social media posting (LinkedIn, Slack, HackerNews)
  - Bluesky integration tests
- Test authentication and authorization flows
- Verify error handling for service failures
- Use real file system operations for any local file handling within these tests

#### 3.3 Configuration Testing
- Test loading of configuration from files using real file system
- Verify fallback to environment variables
- Test configuration validation
- Verify handling of incomplete configurations
- Test configuration serialization

### 4. Test Documentation & Maintenance

#### 4.1 Test Documentation
- Document testing approach for each module
- Create examples for adding new tests
- Document mock implementations and their usage
- Include test coverage reports
- Document the approach of using real file system vs. mocking external APIs

#### 4.2 Test Maintenance
- Ensure tests are maintainable and readable
- Avoid brittle tests that break with minor changes
- Structure tests to mirror the codebase organization
- Create clear naming conventions for tests

## Test Implementation Strategy

### Phase 1: Core Infrastructure
1. Implement test utilities with temporary directory management
2. Set up mock framework for external APIs only
3. Create initial test fixtures

### Phase 2: Basic Unit Tests
1. Implement tests for YAML operations using real file system
2. Add tests for CLI argument handling
3. Create tests for configuration management using real file system

### Phase 3: Feature Tests
1. Implement tests for video management functionality with real file system
2. Create separate test implementations for each external service:
   - YouTube API mock and tests
   - Email service mock and tests
   - Hugo integration mock and tests
   - Social media service mocks and tests
   - Bluesky integration tests
3. Create UI/interaction tests

### Phase 4: Edge Cases & Validation
1. Add tests for error conditions
2. Implement validation tests
3. Test edge cases and boundary conditions

## Test Success Criteria
- All tests pass consistently
- Minimum 80% code coverage
- Tests run without external API dependencies
- Tests complete in a reasonable time
- Edge cases and error conditions are covered

## Test Exclusions
- End-to-end tests requiring actual external services
- Performance testing
- UI appearance testing beyond basic validation
- Tests requiring manual interaction

## Dependencies
- Go testing framework
- httptest package for HTTP mocking
- yaml.v3 for YAML parsing/generation
- Cobra for CLI testing
- os/ioutil and os packages for file system operations
</PRD> 