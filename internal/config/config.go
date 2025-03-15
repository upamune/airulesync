package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the main configuration structure
type Config struct {
	SourceDirs []SourceDir `yaml:"source_dirs" jsonschema:"description=List of source directories containing rule files to be synchronized"`
	TargetDirs []TargetDir `yaml:"target_dirs" jsonschema:"description=List of target directories where rule files will be synchronized to"`
}

// SourceDir represents a source directory configuration
type SourceDir struct {
	Path        string     `yaml:"path" jsonschema:"description=Path to the source directory"`
	Overwrite   *bool      `yaml:"overwrite,omitempty" jsonschema:"description=Whether to overwrite existing files in target directories (default: true)"`
	Files       []FileSpec `yaml:"files" jsonschema:"description=List of files to synchronize from this source directory"`
	IgnoreFiles []string   `yaml:"ignore_files,omitempty" jsonschema:"description=List of file patterns to ignore when synchronizing"`
}

// TargetDir represents a target directory configuration
type TargetDir struct {
	Path        string   `yaml:"path" jsonschema:"description=Path to the target directory"`
	External    bool     `yaml:"external,omitempty" jsonschema:"description=Whether this directory is external to the project (default: false)"`
	IgnoreFiles []string `yaml:"ignore_files,omitempty" jsonschema:"description=List of file patterns to ignore when synchronizing to this target directory"`
}

// FileSpec represents a file specification
type FileSpec struct {
	Pattern     string `yaml:"pattern,omitempty" jsonschema:"description=File pattern to match (glob pattern)"`
	AdjustPaths *bool  `yaml:"adjust_paths,omitempty" jsonschema:"description=Whether to adjust relative paths in the file (default: true)"`
	Overwrite   *bool  `yaml:"overwrite,omitempty" jsonschema:"description=Whether to overwrite existing files (overrides directory setting)"`
}

// UnmarshalYAML implements the yaml.Unmarshaler interface for FileSpec
func (f *FileSpec) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// Try to unmarshal as a string first
	var pattern string
	if err := unmarshal(&pattern); err == nil {
		f.Pattern = pattern
		return nil
	}

	// If that fails, try to unmarshal as a struct
	type fileSpecAlias FileSpec
	return unmarshal((*fileSpecAlias)(f))
}

// GetPattern returns the pattern for the file spec
func (f *FileSpec) GetPattern() string {
	return f.Pattern
}

// ShouldAdjustPaths returns whether paths should be adjusted for this file spec
func (f *FileSpec) ShouldAdjustPaths() bool {
	if f.AdjustPaths == nil {
		return true // Default is true
	}
	return *f.AdjustPaths
}

// ShouldOverwrite returns whether files should be overwritten for this file spec
func (f *FileSpec) ShouldOverwrite(dirDefault bool) bool {
	if f.Overwrite == nil {
		return dirDefault // Use directory default
	}
	return *f.Overwrite
}

// GetDirectoryOverwrite returns the overwrite setting for the source directory
func (s *SourceDir) GetDirectoryOverwrite() bool {
	if s.Overwrite == nil {
		return true // Default is true
	}
	return *s.Overwrite
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if len(c.SourceDirs) == 0 {
		return fmt.Errorf("no source directories specified")
	}

	if len(c.TargetDirs) == 0 {
		return fmt.Errorf("no target directories specified")
	}

	// Validate source directories
	for i, src := range c.SourceDirs {
		if src.Path == "" {
			return fmt.Errorf("source directory %d has no path", i+1)
		}

		if len(src.Files) == 0 {
			return fmt.Errorf("source directory %s has no files specified", src.Path)
		}

		for j, file := range src.Files {
			if file.Pattern == "" {
				return fmt.Errorf("file %d in source directory %s has no pattern", j+1, src.Path)
			}
		}
	}

	// Validate target directories
	for i, tgt := range c.TargetDirs {
		if tgt.Path == "" {
			return fmt.Errorf("target directory %d has no path", i+1)
		}
	}

	return nil
}

// LoadConfig loads the configuration from a file
func LoadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Normalize paths
	for i := range config.SourceDirs {
		config.SourceDirs[i].Path = filepath.Clean(config.SourceDirs[i].Path)
	}

	for i := range config.TargetDirs {
		config.TargetDirs[i].Path = filepath.Clean(config.TargetDirs[i].Path)
	}

	return &config, nil
}

// DefaultConfigPath returns the default configuration path
func DefaultConfigPath() string {
	return ".airulesync.yaml"
}

// SaveConfig saves the configuration to a file
func SaveConfig(config *Config, configPath string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Add header comments
	headerComments := []byte("# yaml-language-server: $schema=https://raw.githubusercontent.com/upamune/airulesync/refs/heads/main/schema.json\n# vim: set ts=2 sw=2 tw=0 fo=cnqoj\n")
	dataWithComments := append(headerComments, data...)

	if err := os.WriteFile(configPath, dataWithComments, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
