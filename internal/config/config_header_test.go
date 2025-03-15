package config

import (
	"os"
	"strings"
	"testing"
)

func TestSaveConfigWithHeaderComments(t *testing.T) {
	// Create a simple config
	cfg := &Config{
		SourceDirs: []SourceDir{
			{
				Path: "test-path",
				Files: []FileSpec{
					{
						Pattern: "test-pattern",
					},
				},
			},
		},
		TargetDirs: []TargetDir{
			{
				Path: "test-target",
			},
		},
	}

	// Create a temporary file
	tempFile := "test-config.yaml"
	defer os.Remove(tempFile)

	// Save the config
	err := SaveConfig(cfg, tempFile)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Read the file
	data, err := os.ReadFile(tempFile)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	// Check for header comments
	content := string(data)
	expectedHeader1 := "# yaml-language-server: $schema=https://raw.githubusercontent.com/upamune/airulesync/refs/heads/main/schema.json"
	expectedHeader2 := "# vim: set ts=2 sw=2 tw=0 fo=cnqoj"

	if !strings.Contains(content, expectedHeader1) {
		t.Errorf("Header comment not found in config file: %s", expectedHeader1)
	}

	if !strings.Contains(content, expectedHeader2) {
		t.Errorf("Header comment not found in config file: %s", expectedHeader2)
	}

	// Check that the YAML content follows the header
	if !strings.Contains(content, "source_dirs:") {
		t.Errorf("YAML content not found in config file")
	}
}
