package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"
)

func TestSchemaGeneration(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run the main function
	main()

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read the captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Verify the output is valid JSON
	var schema map[string]interface{}
	err := json.Unmarshal([]byte(output), &schema)
	if err != nil {
		t.Fatalf("Generated schema is not valid JSON: %v", err)
	}

	// Check for expected fields
	if schema["title"] != "AIRuleSync Configuration Schema" {
		t.Errorf("Schema title is incorrect: %v", schema["title"])
	}

	// Check that property names use YAML naming convention (snake_case)
	defs, ok := schema["$defs"].(map[string]interface{})
	if !ok {
		t.Fatalf("Schema does not have $defs section")
	}

	config, ok := defs["Config"].(map[string]interface{})
	if !ok {
		t.Fatalf("Schema does not have Config definition")
	}

	properties, ok := config["properties"].(map[string]interface{})
	if !ok {
		t.Fatalf("Config does not have properties section")
	}

	// Check for snake_case property names
	for propName := range properties {
		if strings.Contains(propName, "_") == false && propName != strings.ToLower(propName) {
			t.Errorf("Property name %s is not in snake_case format", propName)
		}
	}
}
