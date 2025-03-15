package pathadjust

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAdjustPath(t *testing.T) {
	// Create a temporary directory structure for testing
	tempDir := t.TempDir()

	// Create test directories
	parentDir := filepath.Join(tempDir, "parent")
	childDir := filepath.Join(parentDir, "child")
	siblingDir := filepath.Join(parentDir, "sibling")

	for _, dir := range []string{parentDir, childDir, siblingDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}
	}

	// Create a path adjuster
	adjuster := NewPathAdjuster(false)

	// Test cases for path adjustment
	testCases := []struct {
		name        string
		path        string
		sourceDir   string
		targetDir   string
		expected    string
		shouldError bool
	}{
		{
			name:      "parent to child",
			path:      "./file.txt",
			sourceDir: parentDir,
			targetDir: childDir,
			expected:  "../file.txt",
		},
		{
			name:      "child to parent",
			path:      "./file.txt",
			sourceDir: childDir,
			targetDir: parentDir,
			expected:  "./child/file.txt",
		},
		{
			name:      "sibling to sibling",
			path:      "./file.txt",
			sourceDir: siblingDir,
			targetDir: childDir,
			expected:  "../sibling/file.txt",
		},
		{
			name:      "nested path parent to child",
			path:      "./nested/path/file.txt",
			sourceDir: parentDir,
			targetDir: childDir,
			expected:  "../nested/path/file.txt",
		},
		{
			name:      "nested path child to parent",
			path:      "./nested/path/file.txt",
			sourceDir: childDir,
			targetDir: parentDir,
			expected:  "./child/nested/path/file.txt",
		},
		{
			name:      "parent path to child",
			path:      "../file.txt",
			sourceDir: childDir,
			targetDir: siblingDir,
			expected:  "../file.txt",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			adjusted, err := adjuster.adjustPath(tc.path, tc.sourceDir, tc.targetDir)

			if tc.shouldError {
				if err == nil {
					t.Errorf("Expected error, but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if adjusted != tc.expected {
				t.Errorf("Expected adjusted path '%s', got '%s'", tc.expected, adjusted)
			}
		})
	}
}

func TestAdjustPaths(t *testing.T) {
	// Create a temporary directory structure for testing
	tempDir := t.TempDir()

	// Create test directories
	sourceDir := filepath.Join(tempDir, "source")
	targetDir := filepath.Join(tempDir, "target")

	for _, dir := range []string{sourceDir, targetDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}
	}

	// Create a test file with paths to adjust
	sourceFile := filepath.Join(sourceDir, "test.txt")
	targetFile := filepath.Join(targetDir, "test.txt")

	content := `
This is a test file with paths to adjust.

import "./relative/path/file.js"
import "../another/path/file.js"
require("./module.js")

"path": "./config.json"
file="./data.txt"

[Link](./doc.md)
href="./page.html"
src="./image.png"

"./file.txt"
"../parent.txt"
`

	if err := os.WriteFile(sourceFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Create a path adjuster
	adjuster := NewPathAdjuster(true)

	// Adjust paths in the file
	adjustments, err := adjuster.AdjustPaths(sourceFile, targetFile, sourceDir, targetDir)
	if err != nil {
		t.Fatalf("Failed to adjust paths: %v", err)
	}

	// Verify the adjustments
	if len(adjustments) == 0 {
		t.Errorf("Expected adjustments, but got none")
	}

	// Read the adjusted file
	adjustedContent, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("Failed to read adjusted file: %v", err)
	}

	// Verify that paths were adjusted
	expectedPaths := map[string]string{
		"./relative/path/file.js": "../source/relative/path/file.js",
		// "../another/path/file.js" is not adjusted because it's already a relative path
		"./module.js":   "../source/module.js",
		"./config.json": "../source/config.json",
		"./data.txt":    "../source/data.txt",
		"./doc.md":      "../source/doc.md",
		"./page.html":   "../source/page.html",
		"./image.png":   "../source/image.png",
		"./file.txt":    "../source/file.txt",
		// "../parent.txt" is not adjusted because it's already a relative path
	}

	for original, expected := range expectedPaths {
		found := false
		for _, adj := range adjustments {
			if adj.OriginalPath == original && adj.AdjustedPath == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected adjustment from '%s' to '%s', but not found", original, expected)
		}
	}

	// Verify that the adjusted content contains the expected paths
	adjustedContentStr := string(adjustedContent)
	for _, expected := range expectedPaths {
		if !contains(adjustedContentStr, expected) {
			t.Errorf("Expected adjusted content to contain '%s', but it doesn't", expected)
		}
	}
}

func TestCopyFile(t *testing.T) {
	// Create a temporary directory structure for testing
	tempDir := t.TempDir()

	// Create test directories
	sourceDir := filepath.Join(tempDir, "source")
	targetDir := filepath.Join(tempDir, "target")

	for _, dir := range []string{sourceDir, targetDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}
	}

	// Create a test file to copy
	sourceFile := filepath.Join(sourceDir, "test.txt")
	targetFile := filepath.Join(targetDir, "test.txt")

	content := "This is a test file to copy."

	if err := os.WriteFile(sourceFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Create a path adjuster
	adjuster := NewPathAdjuster(false)

	// Copy the file
	err := adjuster.CopyFile(sourceFile, targetFile)
	if err != nil {
		t.Fatalf("Failed to copy file: %v", err)
	}

	// Verify that the file was copied
	if _, err := os.Stat(targetFile); os.IsNotExist(err) {
		t.Errorf("Target file does not exist")
	}

	// Read the copied file
	copiedContent, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("Failed to read copied file: %v", err)
	}

	// Verify that the content is the same
	if string(copiedContent) != content {
		t.Errorf("Expected copied content '%s', got '%s'", content, string(copiedContent))
	}
}

func TestIsExternalPath(t *testing.T) {
	// Create a path adjuster
	adjuster := NewPathAdjuster(false)

	// Test cases for external path detection
	testCases := []struct {
		path     string
		expected bool
	}{
		{
			path:     "../external",
			expected: true,
		},
		{
			path:     "./internal",
			expected: false,
		},
		{
			path:     "relative",
			expected: false,
		},
		{
			path:     "/absolute",
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.path, func(t *testing.T) {
			result := adjuster.IsExternalPath(tc.path)
			if result != tc.expected {
				t.Errorf("Expected IsExternalPath('%s') to be %v, got %v", tc.path, tc.expected, result)
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && strings.Contains(s, substr)
}
