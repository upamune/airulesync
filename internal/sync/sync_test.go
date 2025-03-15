package sync

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/upamune/airulesync/internal/config"
	"github.com/upamune/airulesync/internal/scanner"
)

func TestSyncFile(t *testing.T) {
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

	// Create a test file to sync
	sourceFile := filepath.Join(sourceDir, ".clinerules")
	content := `
# Test clinerules file
import "./relative/path/file.js"
`

	if err := os.WriteFile(sourceFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Create a test configuration
	cfg := &config.Config{
		SourceDirs: []config.SourceDir{
			{
				Path: sourceDir,
				Files: []config.FileSpec{
					{Pattern: ".clinerules"},
				},
			},
		},
		TargetDirs: []config.TargetDir{
			{
				Path: targetDir,
			},
		},
	}

	// Create a file info
	fileInfo := scanner.FileInfo{
		SourcePath:   sourceFile,
		SourceDir:    sourceDir,
		RelativePath: ".clinerules",
		Pattern:      ".clinerules",
		AdjustPaths:  true,
		Overwrite:    true,
	}

	// Create a syncer
	syncer := NewSyncer(cfg, false, true)

	// Sync the file
	result := syncer.syncFile(fileInfo, cfg.TargetDirs[0])

	// Verify the result
	if !result.Success {
		t.Errorf("Expected sync to succeed, but it failed: %v", result.Error)
	}

	if result.Skipped {
		t.Errorf("Expected file to be synced, but it was skipped: %s", result.SkipReason)
	}

	// Verify that the file was synced
	targetFile := filepath.Join(targetDir, ".clinerules")
	if _, err := os.Stat(targetFile); os.IsNotExist(err) {
		t.Errorf("Target file does not exist")
	}

	// Read the synced file
	syncedContent, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("Failed to read synced file: %v", err)
	}

	// Verify that paths were adjusted
	expectedContent := `
# Test clinerules file
import "../source/relative/path/file.js"
`
	if string(syncedContent) != expectedContent {
		t.Errorf("Expected synced content:\n%s\n\nGot:\n%s", expectedContent, string(syncedContent))
	}
}

func TestSyncFileWithOverwriteFalse(t *testing.T) {
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

	// Create a test file to sync
	sourceFile := filepath.Join(sourceDir, ".clinerules")
	sourceContent := "# Source content"
	if err := os.WriteFile(sourceFile, []byte(sourceContent), 0644); err != nil {
		t.Fatalf("Failed to write source file: %v", err)
	}

	// Create a target file that already exists
	targetFile := filepath.Join(targetDir, ".clinerules")
	targetContent := "# Target content"
	if err := os.WriteFile(targetFile, []byte(targetContent), 0644); err != nil {
		t.Fatalf("Failed to write target file: %v", err)
	}

	// Create a test configuration
	cfg := &config.Config{
		SourceDirs: []config.SourceDir{
			{
				Path: sourceDir,
				Files: []config.FileSpec{
					{Pattern: ".clinerules"},
				},
			},
		},
		TargetDirs: []config.TargetDir{
			{
				Path: targetDir,
			},
		},
	}

	// Create a file info with overwrite=false
	fileInfo := scanner.FileInfo{
		SourcePath:   sourceFile,
		SourceDir:    sourceDir,
		RelativePath: ".clinerules",
		Pattern:      ".clinerules",
		AdjustPaths:  true,
		Overwrite:    false,
	}

	// Create a syncer
	syncer := NewSyncer(cfg, false, true)

	// Sync the file
	result := syncer.syncFile(fileInfo, cfg.TargetDirs[0])

	// Verify the result
	if !result.Skipped {
		t.Errorf("Expected file to be skipped, but it was synced")
	}

	if result.SkipReason != "file exists and overwrite=false" {
		t.Errorf("Expected skip reason 'file exists and overwrite=false', got '%s'", result.SkipReason)
	}

	// Read the target file
	content, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("Failed to read target file: %v", err)
	}

	// Verify that the content was not changed
	if string(content) != targetContent {
		t.Errorf("Expected target content to remain unchanged, but it was modified")
	}
}

func TestSyncFileWithIgnorePattern(t *testing.T) {
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

	// Create a test file to sync
	sourceFile := filepath.Join(sourceDir, ".clinerules")
	content := "# Test clinerules file"
	if err := os.WriteFile(sourceFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Create a test configuration
	cfg := &config.Config{
		SourceDirs: []config.SourceDir{
			{
				Path: sourceDir,
				Files: []config.FileSpec{
					{Pattern: ".clinerules"},
				},
			},
		},
		TargetDirs: []config.TargetDir{
			{
				Path:        targetDir,
				IgnoreFiles: []string{".clinerules"},
			},
		},
	}

	// Create a file info
	fileInfo := scanner.FileInfo{
		SourcePath:   sourceFile,
		SourceDir:    sourceDir,
		RelativePath: ".clinerules",
		Pattern:      ".clinerules",
		AdjustPaths:  true,
		Overwrite:    true,
	}

	// Create a syncer
	syncer := NewSyncer(cfg, false, true)

	// Sync the file
	result := syncer.syncFile(fileInfo, cfg.TargetDirs[0])

	// Verify the result
	if !result.Skipped {
		t.Errorf("Expected file to be skipped, but it was synced")
	}

	if result.SkipReason != "file matches ignore pattern .clinerules in target directory" {
		t.Errorf("Expected skip reason to mention ignore pattern, got '%s'", result.SkipReason)
	}

	// Verify that the file was not synced
	targetFile := filepath.Join(targetDir, ".clinerules")
	if _, err := os.Stat(targetFile); !os.IsNotExist(err) {
		t.Errorf("Target file exists, but it should not")
	}
}

func TestSyncFileDryRun(t *testing.T) {
	// Create a temporary directory structure for testing
	tempDir := t.TempDir()

	// Create test directories
	sourceDir := filepath.Join(tempDir, "source")
	targetDir := filepath.Join(tempDir, "target")

	for _, dir := range []string{sourceDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}
	}

	// Create a test file to sync
	sourceFile := filepath.Join(sourceDir, ".clinerules")
	content := "# Test clinerules file"
	if err := os.WriteFile(sourceFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Create a test configuration
	cfg := &config.Config{
		SourceDirs: []config.SourceDir{
			{
				Path: sourceDir,
				Files: []config.FileSpec{
					{Pattern: ".clinerules"},
				},
			},
		},
		TargetDirs: []config.TargetDir{
			{
				Path: targetDir,
			},
		},
	}

	// Create a file info
	fileInfo := scanner.FileInfo{
		SourcePath:   sourceFile,
		SourceDir:    sourceDir,
		RelativePath: ".clinerules",
		Pattern:      ".clinerules",
		AdjustPaths:  true,
		Overwrite:    true,
	}

	// Create a syncer with dry-run=true
	syncer := NewSyncer(cfg, true, true)

	// Sync the file
	result := syncer.syncFile(fileInfo, cfg.TargetDirs[0])

	// Verify the result
	if !result.Success {
		t.Errorf("Expected dry-run sync to succeed, but it failed: %v", result.Error)
	}

	if result.Skipped {
		t.Errorf("Expected file to be synced in dry-run, but it was skipped: %s", result.SkipReason)
	}

	// Verify that the file was not actually synced
	targetFile := filepath.Join(targetDir, ".clinerules")
	if _, err := os.Stat(targetFile); !os.IsNotExist(err) {
		t.Errorf("Target file exists, but it should not in dry-run mode")
	}
}
