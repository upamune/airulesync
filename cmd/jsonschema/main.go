package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/invopop/jsonschema"
	"github.com/upamune/airulesync/internal/config"
)

func main() {
	// Create a reflector
	r := &jsonschema.Reflector{
		RequiredFromJSONSchemaTags: true,
		FieldNameTag:               "yaml", // Use yaml tag for field names
	}

	// Generate schema from Config struct
	schema := r.Reflect(&config.Config{})

	// Add schema metadata
	schema.Title = "AIRuleSync Configuration Schema"
	schema.Description = "Schema for the AIRuleSync configuration file (.airulesync.yaml)"
	schema.Version = "https://json-schema.org/draft/2020-12/schema"

	// Marshal to JSON with indentation
	jsonData, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling schema to JSON: %v\n", err)
		os.Exit(1)
	}

	// Print to stdout
	fmt.Println(string(jsonData))
}
