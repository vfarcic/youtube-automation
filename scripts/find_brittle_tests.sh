#!/bin/bash
# find_brittle_tests.sh
# This script helps identify potentially brittle tests in the codebase.

set -e

echo "Scanning for potentially brittle tests..."

# Create a temporary directory to store the results
TEMP_DIR=$(mktemp -d)
RESULTS_FILE="$TEMP_DIR/brittle_tests.txt"

# Function to check if file is a Go test file
is_test_file() {
    [[ "$1" == *_test.go ]]
}

# Create report header
echo "Potentially Brittle Test Report" > "$RESULTS_FILE"
echo "=============================" >> "$RESULTS_FILE"
echo "" >> "$RESULTS_FILE"
echo "Generated on: $(date)" >> "$RESULTS_FILE"
echo "" >> "$RESULTS_FILE"

# 1. Find exact string comparisons in assertions
echo "1. Tests with hard-coded string comparisons:" >> "$RESULTS_FILE"
grep -r --include="*_test.go" "assert\.Equal.*\"\(.*\)\"" . | grep -v "assert\.Equal.*t, \"\", " >> "$RESULTS_FILE" || true
echo "" >> "$RESULTS_FILE"

# 2. Find tests that may rely on specific date/time
echo "2. Tests that may rely on specific date/time values:" >> "$RESULTS_FILE"
grep -r --include="*_test.go" -E "(time\.Now|time\.Parse|time\.Unix|time\.Date)" . >> "$RESULTS_FILE" || true
echo "" >> "$RESULTS_FILE"

# 3. Find magic numbers in tests
echo "3. Tests with magic numbers:" >> "$RESULTS_FILE"
grep -r --include="*_test.go" -E "[^A-Za-z0-9\"\._]([-]?[0-9]+)[^A-Za-z0-9\"\._]" . | grep -v "assert\.Equal.*t, [0-9], " >> "$RESULTS_FILE" || true
echo "" >> "$RESULTS_FILE"

# 4. Find files with excessive test setup
echo "4. Tests with potentially excessive setup:" >> "$RESULTS_FILE"
echo "   (Files with more than 30 lines of setup before the first assertion)" >> "$RESULTS_FILE"
for file in $(find . -name "*_test.go"); do
    # Count lines from func Test to first assert
    SETUP_LINES=$(awk '/func Test/ {start=1} /assert\./ {if(start) {print NR; start=0}}' "$file" | head -1)
    if [[ -n "$SETUP_LINES" ]] && [[ "$SETUP_LINES" -gt 30 ]]; then
        echo "   $file: $SETUP_LINES lines of setup" >> "$RESULTS_FILE"
    fi
done
echo "" >> "$RESULTS_FILE"

# 5. Find fragile environment dependencies
echo "5. Tests with potential environment dependencies:" >> "$RESULTS_FILE"
grep -r --include="*_test.go" -E "(os\.Getenv|filepath\.Abs|os\.TempDir)" . >> "$RESULTS_FILE" || true
echo "" >> "$RESULTS_FILE"

# 6. Find tests that may depend on execution order
echo "6. Tests that may depend on execution order (no t.Parallel):" >> "$RESULTS_FILE"
for file in $(find . -name "*_test.go"); do
    if grep -q "func Test" "$file" && ! grep -q "t\.Parallel" "$file"; then
        echo "   $file: No parallel test execution" >> "$RESULTS_FILE"
    fi
done

# Print the final report
cat "$RESULTS_FILE"

# Save the report to a file in the current directory
cp "$RESULTS_FILE" "./brittle_tests_report.txt"
echo "Report saved to: ./brittle_tests_report.txt"

# Cleanup
rm -rf "$TEMP_DIR"
echo "Scan complete." 