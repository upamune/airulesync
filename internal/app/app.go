package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/upamune/airulesync/internal/config"
	"github.com/upamune/airulesync/internal/scanner"
	"github.com/upamune/airulesync/internal/sync"
	"github.com/upamune/airulesync/internal/version"
)

// App represents the application
type App struct {
	ConfigPath string
	Verbose    bool
}

// NewApp creates a new application
func NewApp(configPath string, verbose bool) *App {
	return &App{
		ConfigPath: configPath,
		Verbose:    verbose,
	}
}

// RunSync runs the sync command
func (a *App) RunSync(dryRun bool) error {
	// Load configuration
	cfg, err := config.LoadConfig(a.ConfigPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create a syncer
	syncer := sync.NewSyncer(cfg, dryRun, a.Verbose)

	// Run the synchronization
	report, err := syncer.Sync()
	if err != nil {
		return fmt.Errorf("synchronization failed: %w", err)
	}

	// Print the report
	syncer.PrintReport(report, dryRun)

	return nil
}

// RunInit runs the init command
func (a *App) RunInit(dir string) error {
	// If no directory is specified, use the current directory
	if dir == "" {
		var err error
		dir, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
	}

	// Ensure the directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("directory %s does not exist", dir)
	}

	// Check if configuration file already exists
	configPath := config.DefaultConfigPath()
	if _, err := os.Stat(configPath); err == nil {
		fmt.Printf("Configuration file %s already exists. Skipping initialization.\n", configPath)
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to check if configuration file exists: %w", err)
	}

	fmt.Printf("Scanning directory for rule files...\n")

	// Create a scanner
	s := scanner.NewScanner(nil)

	// Scan the directory for rule files
	ruleFiles, err := s.ScanDirectory(dir)
	if err != nil {
		return fmt.Errorf("failed to scan directory: %w", err)
	}

	var cfg *config.Config

	if len(ruleFiles) == 0 {
		fmt.Println("No rule files found in the directory.")
		// Create a default empty configuration
		cfg = &config.Config{
			SourceDirs: []config.SourceDir{},
			TargetDirs: []config.TargetDir{},
		}
	} else {
		fmt.Printf("Found %d potential rule files:\n", len(ruleFiles))
		for _, file := range ruleFiles {
			fmt.Printf("- %s\n", filepath.Join(dir, file))
		}

		// Find potential target directories
		fmt.Println("\nDetecting potential target directories...")
		targetDirs, err := s.FindPotentialTargetDirs(dir)
		if err != nil {
			return fmt.Errorf("failed to find potential target directories: %w", err)
		}

		if len(targetDirs) == 0 {
			fmt.Println("No potential target directories found.")
		} else {
			fmt.Printf("Detected potential target directories:\n")
			for _, targetDir := range targetDirs {
				fmt.Printf("- %s\n", filepath.Join(dir, targetDir))
			}
		}

		// Generate a configuration
		cfg = a.generateConfig(dir, ruleFiles, targetDirs)
	}

	// Save the configuration
	if err := config.SaveConfig(cfg, configPath); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Printf("\nConfiguration written to %s\n", configPath)
	fmt.Println("Review and edit the configuration as needed before running 'airulesync sync'")

	return nil
}

// generateConfig generates a configuration based on the scan results
func (a *App) generateConfig(baseDir string, ruleFiles, targetDirs []string) *config.Config {
	// Group rule files by directory
	filesByDir := make(map[string][]string)
	hasCursorRules := false

	for _, file := range ruleFiles {
		dir := filepath.Dir(file)
		if dir == "." {
			dir = ""
		}

		// Check if this is a file in the .cursor/rules directory
		if strings.HasPrefix(dir, ".cursor/rules") {
			hasCursorRules = true
			continue // Skip individual .cursor/rules files
		}

		filesByDir[dir] = append(filesByDir[dir], filepath.Base(file))
	}

	// Create source directories
	var sourceDirs []config.SourceDir

	// Create a source directory for the base directory
	var baseFileSpecs []config.FileSpec

	// If .cursor/rules directory exists, add a pattern for all .mdc files
	if hasCursorRules {
		baseFileSpecs = append(baseFileSpecs, config.FileSpec{
			Pattern: ".cursor/rules/*.mdc",
		})
	}

	// Add other files from the base directory
	for dir, files := range filesByDir {
		for _, file := range files {
			// Check if the file is in a subdirectory
			if dir != "" {
				baseFileSpecs = append(baseFileSpecs, config.FileSpec{
					Pattern: filepath.Join(dir, file),
				})
			} else {
				baseFileSpecs = append(baseFileSpecs, config.FileSpec{
					Pattern: file,
				})
			}
		}
	}

	// Add the source directory if it has files
	if len(baseFileSpecs) > 0 {
		sourceDirs = append(sourceDirs, config.SourceDir{
			Path:  ".", // Use relative path instead of absolute path
			Files: baseFileSpecs,
		})
	}

	// Create empty target directories slice
	// This allows users to manually configure target directories as needed
	var targetDirConfigs []config.TargetDir

	return &config.Config{
		SourceDirs: sourceDirs,
		TargetDirs: targetDirConfigs, // Empty slice
	}
}

// RunVersion runs the version command
func (a *App) RunVersion() error {
	fmt.Println(version.FormatBuildInfo())
	return nil
}
