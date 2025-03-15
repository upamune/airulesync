package config

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	// Write a valid configuration to the file
	validConfig := `
source_dirs:
  - path: "./src/main-project"
    files:
      - ".clinerules"
      - pattern: ".cursor/rules/**/*.mdc"
        adjust_paths: true
      - pattern: ".roomodes"
        adjust_paths: false
    ignore_files:
      - ".cursor/rules/private/*.mdc"

target_dirs:
  - path: "./src/sub-project-a"
  - path: "./src/sub-project-b"
  - path: "../other-repo/src/component"
    external: true
`
	err := os.WriteFile(configPath, []byte(validConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	// Load the configuration
	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load valid config: %v", err)
	}

	// Verify the configuration
	if len(cfg.SourceDirs) != 1 {
		t.Errorf("Expected 1 source directory, got %d", len(cfg.SourceDirs))
	}

	if len(cfg.TargetDirs) != 3 {
		t.Errorf("Expected 3 target directories, got %d", len(cfg.TargetDirs))
	}

	// Check source directory
	sourceDir := cfg.SourceDirs[0]
	if sourceDir.Path != "src/main-project" {
		t.Errorf("Expected source path 'src/main-project', got '%s'", sourceDir.Path)
	}

	if len(sourceDir.Files) != 3 {
		t.Errorf("Expected 3 files in source directory, got %d", len(sourceDir.Files))
	}

	if len(sourceDir.IgnoreFiles) != 1 {
		t.Errorf("Expected 1 ignore file in source directory, got %d", len(sourceDir.IgnoreFiles))
	}

	// Check file specs
	if sourceDir.Files[0].Pattern != ".clinerules" {
		t.Errorf("Expected first file pattern '.clinerules', got '%s'", sourceDir.Files[0].Pattern)
	}

	if sourceDir.Files[1].Pattern != ".cursor/rules/**/*.mdc" {
		t.Errorf("Expected second file pattern '.cursor/rules/**/*.mdc', got '%s'", sourceDir.Files[1].Pattern)
	}

	if !sourceDir.Files[1].ShouldAdjustPaths() {
		t.Errorf("Expected second file to adjust paths")
	}

	if sourceDir.Files[2].ShouldAdjustPaths() {
		t.Errorf("Expected third file to not adjust paths")
	}

	// Check target directories
	if cfg.TargetDirs[0].Path != "src/sub-project-a" {
		t.Errorf("Expected first target path 'src/sub-project-a', got '%s'", cfg.TargetDirs[0].Path)
	}

	if cfg.TargetDirs[1].Path != "src/sub-project-b" {
		t.Errorf("Expected second target path 'src/sub-project-b', got '%s'", cfg.TargetDirs[1].Path)
	}

	if cfg.TargetDirs[2].Path != "../other-repo/src/component" {
		t.Errorf("Expected third target path '../other-repo/src/component', got '%s'", cfg.TargetDirs[2].Path)
	}

	if !cfg.TargetDirs[2].External {
		t.Errorf("Expected third target to be external")
	}
}

func TestLoadInvalidConfig(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "invalid-config.yaml")

	// Test cases for invalid configurations
	testCases := []struct {
		name   string
		config string
	}{
		{
			name: "missing source directories",
			config: `
target_dirs:
  - path: "./src/sub-project-a"
`,
		},
		{
			name: "missing target directories",
			config: `
source_dirs:
  - path: "./src/main-project"
    files:
      - ".clinerules"
`,
		},
		{
			name: "empty source path",
			config: `
source_dirs:
  - path: ""
    files:
      - ".clinerules"
target_dirs:
  - path: "./src/sub-project-a"
`,
		},
		{
			name: "empty target path",
			config: `
source_dirs:
  - path: "./src/main-project"
    files:
      - ".clinerules"
target_dirs:
  - path: ""
`,
		},
		{
			name: "no files in source directory",
			config: `
source_dirs:
  - path: "./src/main-project"
    files: []
target_dirs:
  - path: "./src/sub-project-a"
`,
		},
		{
			name: "empty file pattern",
			config: `
source_dirs:
  - path: "./src/main-project"
    files:
      - ""
target_dirs:
  - path: "./src/sub-project-a"
`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Write the invalid configuration to the file
			err := os.WriteFile(configPath, []byte(tc.config), 0644)
			if err != nil {
				t.Fatalf("Failed to write test config file: %v", err)
			}

			// Try to load the configuration
			_, err = LoadConfig(configPath)
			if err == nil {
				t.Errorf("Expected error for invalid config '%s', but got nil", tc.name)
			}
		})
	}
}

func TestFileSpecUnmarshalYAML(t *testing.T) {
	// Test cases for FileSpec unmarshaling
	testCases := []struct {
		name            string
		yaml            string
		expectedPattern string
		adjustPaths     bool
		overwrite       bool
	}{
		{
			name:            "string pattern",
			yaml:            `".clinerules"`,
			expectedPattern: ".clinerules",
			adjustPaths:     true, // default
			overwrite:       true, // default
		},
		{
			name: "struct pattern with adjust_paths true",
			yaml: `
pattern: ".cursor/rules/**/*.mdc"
adjust_paths: true
`,
			expectedPattern: ".cursor/rules/**/*.mdc",
			adjustPaths:     true,
			overwrite:       true, // default
		},
		{
			name: "struct pattern with adjust_paths false",
			yaml: `
pattern: ".roomodes"
adjust_paths: false
`,
			expectedPattern: ".roomodes",
			adjustPaths:     false,
			overwrite:       true, // default
		},
		{
			name: "struct pattern with overwrite false",
			yaml: `
pattern: ".cursorignore"
overwrite: false
`,
			expectedPattern: ".cursorignore",
			adjustPaths:     true, // default
			overwrite:       false,
		},
		{
			name: "struct pattern with all options",
			yaml: `
pattern: ".rooignore"
adjust_paths: false
overwrite: false
`,
			expectedPattern: ".rooignore",
			adjustPaths:     false,
			overwrite:       false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a temporary YAML file
			tempDir := t.TempDir()
			yamlPath := filepath.Join(tempDir, "file-spec.yaml")

			// Write the YAML to the file
			err := os.WriteFile(yamlPath, []byte(tc.yaml), 0644)
			if err != nil {
				t.Fatalf("Failed to write test YAML file: %v", err)
			}

			// Read the YAML file
			data, err := os.ReadFile(yamlPath)
			if err != nil {
				t.Fatalf("Failed to read test YAML file: %v", err)
			}

			// Unmarshal the YAML
			var fileSpec FileSpec
			err = yaml.Unmarshal(data, &fileSpec)
			if err != nil {
				t.Fatalf("Failed to unmarshal YAML: %v", err)
			}

			// Verify the FileSpec
			if fileSpec.Pattern != tc.expectedPattern {
				t.Errorf("Expected pattern '%s', got '%s'", tc.expectedPattern, fileSpec.Pattern)
			}

			if fileSpec.ShouldAdjustPaths() != tc.adjustPaths {
				t.Errorf("Expected adjust_paths %v, got %v", tc.adjustPaths, fileSpec.ShouldAdjustPaths())
			}

			if fileSpec.ShouldOverwrite(true) != tc.overwrite {
				t.Errorf("Expected overwrite %v, got %v", tc.overwrite, fileSpec.ShouldOverwrite(true))
			}
		})
	}
}
