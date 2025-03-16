package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunSync(t *testing.T) {
	// Skip this test in short mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Copy the test fixtures to the temporary directory
	fixturesDir := "../../test/fixtures/test-project"
	targetDir := filepath.Join(tempDir, "test-project")

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		t.Fatalf("Failed to create target directory: %v", err)
	}

	// Copy the .clinerules file
	sourceFile := filepath.Join(fixturesDir, ".clinerules")
	targetFile := filepath.Join(targetDir, ".clinerules")
	if err := copyFile(sourceFile, targetFile); err != nil {
		t.Fatalf("Failed to copy .clinerules file: %v", err)
	}

	// Copy the .airulesync.yaml file
	sourceConfig := filepath.Join(fixturesDir, ".airulesync.yaml")
	targetConfig := filepath.Join(targetDir, ".airulesync.yaml")
	if err := copyFile(sourceConfig, targetConfig); err != nil {
		t.Fatalf("Failed to copy .airulesync.yaml file: %v", err)
	}

	// Create the target directories
	subProjectA := filepath.Join(targetDir, "sub-project-a")
	subProjectB := filepath.Join(targetDir, "sub-project-b")
	for _, dir := range []string{subProjectA, subProjectB} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create sub-project directory: %v", err)
		}
	}

	// Change to the target directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(targetDir); err != nil {
		t.Fatalf("Failed to change to target directory: %v", err)
	}

	// Create the application
	app := NewApp(".airulesync.yaml", true)

	// Run the sync command
	if err := app.RunSync(false); err != nil {
		t.Fatalf("Failed to run sync command: %v", err)
	}

	// Verify that the files were synced
	subProjectAFile := filepath.Join(subProjectA, ".clinerules")
	if _, err := os.Stat(subProjectAFile); os.IsNotExist(err) {
		t.Errorf("File was not synced to sub-project-a")
	}

	subProjectBFile := filepath.Join(subProjectB, ".clinerules")
	if _, err := os.Stat(subProjectBFile); os.IsNotExist(err) {
		t.Errorf("File was not synced to sub-project-b")
	}

	// Read the synced files
	subProjectAContent, err := os.ReadFile(subProjectAFile)
	if err != nil {
		t.Fatalf("Failed to read synced file in sub-project-a: %v", err)
	}

	subProjectBContent, err := os.ReadFile(subProjectBFile)
	if err != nil {
		t.Fatalf("Failed to read synced file in sub-project-b: %v", err)
	}

	// Verify that paths were adjusted
	expectedPathA := "../relative/path/file.js"
	if !contains(string(subProjectAContent), expectedPathA) {
		t.Errorf("Expected path '%s' in sub-project-a file, but it wasn't found", expectedPathA)
	}

	expectedPathB := "../relative/path/file.js"
	if !contains(string(subProjectBContent), expectedPathB) {
		t.Errorf("Expected path '%s' in sub-project-b file, but it wasn't found", expectedPathB)
	}
}

func TestRunInit(t *testing.T) {
	// Skip this test in short mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create a test project structure
	projectDir := filepath.Join(tempDir, "init-test")
	subDirA := filepath.Join(projectDir, "sub-a")
	subDirB := filepath.Join(projectDir, "sub-b")
	cursorRulesDir := filepath.Join(projectDir, ".cursor", "rules")

	for _, dir := range []string{projectDir, subDirA, subDirB, cursorRulesDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}
	}

	// Create test files
	files := map[string]string{
		filepath.Join(projectDir, ".clinerules"):   "# Test clinerules file",
		filepath.Join(projectDir, ".roomodes"):     "# Test roomodes file",
		filepath.Join(cursorRulesDir, "rule1.mdc"): "# Test rule1 file",
		filepath.Join(subDirA, "main.go"):          "package main",
		filepath.Join(subDirB, "app.js"):           "console.log('Hello');",
	}

	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}
	}

	// Change to the temporary directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temporary directory: %v", err)
	}

	// Create the application
	app := NewApp(".airulesync.yaml", true)

	// Run the init command
	if err := app.RunInit(projectDir); err != nil {
		t.Fatalf("Failed to run init command: %v", err)
	}

	// Verify that the configuration file was created
	configPath := filepath.Join(tempDir, ".airulesync.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("Configuration file was not created")
	}

	// Read the configuration file
	configContent, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read configuration file: %v", err)
	}

	// Verify that the configuration contains the expected content
	expectedContent := []string{
		"source_dirs:",
		"path: .", // Now using relative path
		".clinerules",
		".roomodes",
		".cursor/rules/*.mdc",
		"target_dirs: []", // Now empty target_dirs
	}

	for _, expected := range expectedContent {
		if !contains(string(configContent), expected) {
			t.Errorf("Expected configuration to contain '%s', but it wasn't found", expected)
		}
	}
}

func TestRunInitWithExistingConfig(t *testing.T) {
	// Skip this test in short mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create a test project structure
	projectDir := filepath.Join(tempDir, "init-test-existing")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create a test file
	testFile := filepath.Join(projectDir, ".clinerules")
	if err := os.WriteFile(testFile, []byte("# Test clinerules file"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Change to the temporary directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temporary directory: %v", err)
	}

	// Create an existing configuration file
	existingConfig := `# yaml-language-server: $schema=https://raw.githubusercontent.com/upamune/airulesync/refs/heads/main/schema.json
# vim: set ts=2 sw=2 tw=0 fo=cnqoj
source_dirs:
- path: .
  files:
  - pattern: existing-pattern.txt
target_dirs:
- path: existing-target
`
	configPath := filepath.Join(tempDir, ".airulesync.yaml")
	if err := os.WriteFile(configPath, []byte(existingConfig), 0644); err != nil {
		t.Fatalf("Failed to write existing config file: %v", err)
	}

	// Create the application
	app := NewApp(".airulesync.yaml", true)

	// Run the init command
	if err := app.RunInit(projectDir); err != nil {
		t.Fatalf("Failed to run init command: %v", err)
	}

	// Read the configuration file
	configContent, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read configuration file: %v", err)
	}

	// Verify that the configuration was not overwritten
	if !contains(string(configContent), "existing-pattern.txt") {
		t.Errorf("Expected configuration to contain original content, but it was overwritten")
	}
	if !contains(string(configContent), "existing-target") {
		t.Errorf("Expected configuration to contain original target directory, but it was overwritten")
	}
}

func TestRunInitWithNoRuleFiles(t *testing.T) {
	// Skip this test in short mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create a test project structure with no rule files
	projectDir := filepath.Join(tempDir, "init-test-empty")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create some non-rule files
	files := map[string]string{
		filepath.Join(projectDir, "main.go"):   "package main",
		filepath.Join(projectDir, "README.md"): "# Test Project",
	}

	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}
	}

	// Change to the temporary directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temporary directory: %v", err)
	}

	// Create the application
	app := NewApp(".airulesync.yaml", true)

	// Run the init command
	if err := app.RunInit(projectDir); err != nil {
		t.Fatalf("Failed to run init command: %v", err)
	}

	// Verify that the configuration file was created
	configPath := filepath.Join(tempDir, ".airulesync.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("Configuration file was not created")
	}

	// Read the configuration file
	configContent, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read configuration file: %v", err)
	}

	// Verify that the configuration contains the expected default content
	expectedContent := []string{
		"# yaml-language-server: $schema=https://raw.githubusercontent.com/upamune/airulesync/refs/heads/main/schema.json",
		"# vim: set ts=2 sw=2 tw=0 fo=cnqoj",
		"source_dirs: []",
		"target_dirs: []",
	}

	for _, expected := range expectedContent {
		if !contains(string(configContent), expected) {
			t.Errorf("Expected configuration to contain '%s', but it wasn't found", expected)
		}
	}
}

// Helper function to copy a file
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, data, 0644)
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && strings.Contains(s, substr)
}
