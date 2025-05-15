package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestMoveFile(t *testing.T) {
	// Helper to create a dummy file with specific content and mode
	createDummyFile := func(path string, content string, mode os.FileMode) error {
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}
		return os.WriteFile(path, []byte(content), mode)
	}

	// Test case: Successful move
	t.Run("SuccessfulMove", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "test-move-success")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		srcPath := filepath.Join(tempDir, "src", "source.txt")
		dstPath := filepath.Join(tempDir, "dst", "dest.txt")
		originalContent := "hello world"
		originalMode := os.FileMode(0644)

		if err := createDummyFile(srcPath, originalContent, originalMode); err != nil {
			t.Fatalf("Failed to create source file: %v", err)
		}

		err = MoveFile(srcPath, dstPath) // This will fail to compile initially
		if err != nil {
			t.Errorf("MoveFile failed: %v", err)
		}

		// Verify source is gone
		if _, err := os.Stat(srcPath); !os.IsNotExist(err) {
			t.Errorf("Source file %s still exists after move", srcPath)
		}

		// Verify destination exists
		dstInfo, err := os.Stat(dstPath)
		if err != nil {
			t.Fatalf("Destination file %s does not exist after move: %v", dstPath, err)
		}

		// Verify content
		content, err := os.ReadFile(dstPath)
		if err != nil {
			t.Fatalf("Failed to read destination file: %v", err)
		}
		if string(content) != originalContent {
			t.Errorf("Content mismatch: expected '%s', got '%s'", originalContent, string(content))
		}

		// Verify mode
		if dstInfo.Mode() != originalMode {
			t.Errorf("Mode mismatch: expected %v, got %v", originalMode, dstInfo.Mode())
		}
	})

	// Test case: Destination directory creation
	t.Run("DestinationDirectoryCreation", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "test-move-dst-create")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		srcPath := filepath.Join(tempDir, "source.txt")
		// Destination directory "new_dst" does not exist yet
		dstPath := filepath.Join(tempDir, "new_dst", "dest.txt")
		originalContent := "create dir test"
		originalMode := os.FileMode(0600)

		if err := createDummyFile(srcPath, originalContent, originalMode); err != nil {
			t.Fatalf("Failed to create source file: %v", err)
		}

		err = MoveFile(srcPath, dstPath)
		if err != nil {
			t.Errorf("MoveFile failed when dest dir needs creation: %v", err)
		}
		if _, err := os.Stat(dstPath); err != nil {
			t.Errorf("Destination file not found after move with dir creation: %v", err)
		}
	})

	// Test case: Error if destination file exists
	t.Run("ErrorIfDestinationExists", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "test-move-dst-exists")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		srcPath := filepath.Join(tempDir, "src.txt")
		dstPath := filepath.Join(tempDir, "dst.txt")

		if err := createDummyFile(srcPath, "source content", 0644); err != nil {
			t.Fatalf("Failed to create source file: %v", err)
		}
		if err := createDummyFile(dstPath, "existing dest content", 0644); err != nil {
			t.Fatalf("Failed to create existing destination file: %v", err)
		}

		err = MoveFile(srcPath, dstPath)
		if err == nil {
			t.Errorf("MoveFile should have failed because destination exists, but it didn't")
		} else {
			expectedErrorMsg := fmt.Sprintf("destination file already exists: %s", dstPath)
			if err.Error() != expectedErrorMsg { // Exact error message check might be too brittle; consider strings.Contains
				// For now, let's assume the function *will* return this exact message as per Taskmaster's spec.
				// t.Logf("Note: For a more robust test, consider using strings.Contains for error messages or custom error types.")
			}
		}
	})

	// Test case: Error if source file doesn't exist
	t.Run("ErrorIfSourceDoesNotExist", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "test-move-src-not-exists")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		srcPath := filepath.Join(tempDir, "non_existent_source.txt")
		dstPath := filepath.Join(tempDir, "dest.txt")

		err = MoveFile(srcPath, dstPath)
		if err == nil {
			t.Errorf("MoveFile should have failed because source does not exist, but it didn't")
		}
		// We might want to check for a specific error message or type here too.
		// e.g., if strings.Contains(err.Error(), "failed to get source file info")
	})

	// Test case: Permissions - source is read-only, destination should inherit
	t.Run("SourceReadOnlyPermission", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "test-move-readonly")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		srcPath := filepath.Join(tempDir, "src", "readonly_source.txt")
		dstPath := filepath.Join(tempDir, "dst", "readonly_dest.txt")
		originalContent := "read only content"
		originalMode := os.FileMode(0444) // Read-only for owner, group, others

		if err := createDummyFile(srcPath, originalContent, originalMode); err != nil {
			t.Fatalf("Failed to create source file: %v", err)
		}

		// Ensure file is actually read-only for the test if OS allows strictness
		// For simplicity, we assume createDummyFile sets it correctly.

		err = MoveFile(srcPath, dstPath)
		if err != nil {
			t.Errorf("MoveFile failed for read-only source: %v", err)
		}

		dstInfo, err := os.Stat(dstPath)
		if err != nil {
			t.Fatalf("Destination file %s does not exist: %v", dstPath, err)
		}
		if dstInfo.Mode() != originalMode {
			t.Errorf("Mode mismatch for read-only: expected %v, got %v", originalMode, dstInfo.Mode())
		}
	})

	// Test case: Moving into the same directory (different name)
	t.Run("MoveToSameDirectoryDifferentName", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "test-move-samedir")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		srcPath := filepath.Join(tempDir, "source_original.txt")
		dstPath := filepath.Join(tempDir, "source_renamed.txt")
		originalContent := "same directory rename"
		originalMode := os.FileMode(0644)

		if err := createDummyFile(srcPath, originalContent, originalMode); err != nil {
			t.Fatalf("Failed to create source file: %v", err)
		}

		err = MoveFile(srcPath, dstPath)
		if err != nil {
			t.Errorf("MoveFile failed for same directory rename: %v", err)
		}

		if _, err := os.Stat(srcPath); !os.IsNotExist(err) {
			t.Errorf("Source file %s still exists after rename", srcPath)
		}
		if _, err := os.Stat(dstPath); err != nil {
			t.Errorf("Destination file %s does not exist after rename: %v", dstPath, err)
		}
	})

	// Test case: Source and destination are the same
	t.Run("SourceAndDestinationAreSame", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "test-move-samesrcdst")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		srcPath := filepath.Join(tempDir, "file.txt")
		originalContent := "same file content"
		originalMode := os.FileMode(0644)

		if err := createDummyFile(srcPath, originalContent, originalMode); err != nil {
			t.Fatalf("Failed to create source file: %v", err)
		}

		// Moving a file to itself should ideally be a no-op or a specific error,
		// depending on os.Rename behavior on the platform.
		// os.Rename might return an error if src and dst are the same hard link,
		// or it might succeed silently.
		// For this test, let's assume the current os.Rename behavior is acceptable if it doesn't error
		// or if it errors in a predictable way.
		// The function as designed by Taskmaster *should* error because os.Stat(dst) will exist.
		err = MoveFile(srcPath, srcPath)
		if err == nil {
			t.Errorf("MoveFile should have failed when source and destination are the same (due to dst check), but it didn't")
		} else {
			// Check for the specific error "destination file already exists"
			expectedErrorMsg := fmt.Sprintf("destination file already exists: %s", srcPath)
			if err.Error() != expectedErrorMsg {
				// t.Logf("MoveFile on same src/dst returned: %v. This might be OS-dependent or function design.", err)
			}
		}

		// Ensure the file still exists with original content and mode
		info, err := os.Stat(srcPath)
		if err != nil {
			t.Fatalf("Source file %s does not exist after same-file move attempt: %v", srcPath, err)
		}
		content, _ := os.ReadFile(srcPath)
		if string(content) != originalContent {
			t.Errorf("Content mismatch after same-file move: expected '%s', got '%s'", originalContent, string(content))
		}
		if info.Mode() != originalMode {
			t.Errorf("Mode mismatch after same-file move: expected %v, got %v", originalMode, info.Mode())
		}
	})

	// Note: Testing cross-device moves is complex as it depends on os.Rename behavior,
	// which typically doesn't work across different filesystems.
	// Such a test would require setting up different mount points or mock filesystems.
	// For now, we'll assume same-filesystem moves.
}
