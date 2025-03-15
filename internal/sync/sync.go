package sync

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/upamune/airulesync/internal/config"
	"github.com/upamune/airulesync/internal/pathadjust"
	"github.com/upamune/airulesync/internal/scanner"
)

// SyncResult represents the result of a synchronization operation
type SyncResult struct {
	SourceFile      string
	TargetFile      string
	Success         bool
	Error           error
	PathAdjustments []pathadjust.AdjustmentResult
	Skipped         bool
	SkipReason      string
}

// SyncReport represents a report of all synchronization operations
type SyncReport struct {
	Results []SyncResult
}

// Syncer is responsible for synchronizing files between directories
type Syncer struct {
	Config       *config.Config
	Scanner      *scanner.Scanner
	PathAdjuster *pathadjust.PathAdjuster
	DryRun       bool
	Verbose      bool
}

// NewSyncer creates a new syncer
func NewSyncer(cfg *config.Config, dryRun, verbose bool) *Syncer {
	return &Syncer{
		Config:       cfg,
		Scanner:      scanner.NewScanner(cfg),
		PathAdjuster: pathadjust.NewPathAdjuster(verbose),
		DryRun:       dryRun,
		Verbose:      verbose,
	}
}

// Sync synchronizes files between directories
func (s *Syncer) Sync() (*SyncReport, error) {
	// Scan source directories for files to synchronize
	files, err := s.Scanner.ScanSourceDirs()
	if err != nil {
		return nil, fmt.Errorf("failed to scan source directories: %w", err)
	}

	// Synchronize each file to each target directory
	var results []SyncResult
	for _, file := range files {
		for _, targetDir := range s.Config.TargetDirs {
			result := s.syncFile(file, targetDir)
			results = append(results, result)
		}
	}

	return &SyncReport{
		Results: results,
	}, nil
}

// syncFile synchronizes a single file to a target directory
func (s *Syncer) syncFile(file scanner.FileInfo, targetDir config.TargetDir) SyncResult {
	// Calculate the target file path
	relPath := file.RelativePath
	targetPath := filepath.Join(targetDir.Path, relPath)

	// Create a result object
	result := SyncResult{
		SourceFile: file.SourcePath,
		TargetFile: targetPath,
		Success:    false,
		Skipped:    false,
	}

	// Check if the file should be ignored
	for _, ignorePattern := range targetDir.IgnoreFiles {
		if match, _ := filepath.Match(ignorePattern, relPath); match {
			result.Skipped = true
			result.SkipReason = fmt.Sprintf("file matches ignore pattern %s in target directory", ignorePattern)
			return result
		}
	}

	// Check if the target file exists and should be overwritten
	if !file.Overwrite {
		if _, err := os.Stat(targetPath); err == nil {
			result.Skipped = true
			result.SkipReason = "file exists and overwrite=false"
			return result
		}
	}

	// If this is a dry run, just return the result
	if s.DryRun {
		result.Success = true
		return result
	}

	// Ensure the target directory exists
	targetDirPath := filepath.Dir(targetPath)
	if err := os.MkdirAll(targetDirPath, 0755); err != nil {
		result.Error = fmt.Errorf("failed to create target directory: %w", err)
		return result
	}

	// Synchronize the file
	if file.AdjustPaths {
		// Adjust paths in the file
		adjustments, err := s.PathAdjuster.AdjustPaths(
			file.SourcePath,
			targetPath,
			file.SourceDir,
			targetDir.Path,
		)
		if err != nil {
			result.Error = fmt.Errorf("failed to adjust paths: %w", err)
			return result
		}
		result.PathAdjustments = adjustments
	} else {
		// Copy the file without adjusting paths
		if err := s.PathAdjuster.CopyFile(file.SourcePath, targetPath); err != nil {
			result.Error = fmt.Errorf("failed to copy file: %w", err)
			return result
		}
	}

	result.Success = true
	return result
}

// PrintReport prints a report of the synchronization operations
func (s *Syncer) PrintReport(report *SyncReport, dryRun bool) {
	prefix := ""
	if dryRun {
		prefix = "[DRY-RUN] "
	}

	fmt.Printf("%sStarting synchronization process\n", prefix)
	fmt.Printf("%sScanning source directories for target files...\n", prefix)

	// Group results by source file for better readability
	sourceFiles := make(map[string][]SyncResult)
	for _, result := range report.Results {
		sourceFiles[result.SourceFile] = append(sourceFiles[result.SourceFile], result)
	}

	// Print files to synchronize
	fmt.Printf("\n%sFiles to synchronize:\n", prefix)
	syncCount := 0
	skipCount := 0

	for sourceFile, results := range sourceFiles {
		for _, result := range results {
			if !result.Skipped {
				syncCount++
				fmt.Printf("%s- '%s' -> '%s'\n", prefix, sourceFile, result.TargetFile)

				if result.PathAdjustments != nil && len(result.PathAdjustments) > 0 {
					fmt.Printf("%s  * Path adjustments: %d locations\n", prefix, len(result.PathAdjustments))

					if s.Verbose {
						for _, adj := range result.PathAdjustments {
							fmt.Printf("%s    - Line %d: '%s' -> '%s'\n", prefix, adj.LineNumber, adj.OriginalPath, adj.AdjustedPath)
						}
					}
				} else if result.PathAdjustments != nil {
					fmt.Printf("%s  * Path adjustments: 0 locations\n", prefix)
				} else {
					fmt.Printf("%s  * No path adjustment (as configured)\n", prefix)
				}

				// Check if this is a cross-repository sync
				if s.PathAdjuster.IsExternalPath(filepath.Dir(result.TargetFile)) {
					fmt.Printf("%s  * Warning: Cross-repository paths may require manual verification\n", prefix)
				}
			}
		}
	}

	// Print files to skip
	fmt.Printf("\n%sFiles to skip:\n", prefix)
	for sourceFile, results := range sourceFiles {
		for _, result := range results {
			if result.Skipped {
				skipCount++
				fmt.Printf("%s- '%s' -> '%s' (%s)\n", prefix, sourceFile, result.TargetFile, result.SkipReason)
			}
		}
	}

	// Print summary
	fmt.Printf("\n%sSynchronization process completed\n", prefix)
	fmt.Printf("%s- Files synchronized: %d\n", prefix, syncCount)
	fmt.Printf("%s- Files skipped: %d\n", prefix, skipCount)

	// Print errors if any
	errorCount := 0
	for _, result := range report.Results {
		if result.Error != nil {
			errorCount++
		}
	}

	if errorCount > 0 {
		fmt.Printf("%s- Errors encountered: %d\n", prefix, errorCount)

		if s.Verbose {
			fmt.Printf("\n%sErrors:\n", prefix)
			for _, result := range report.Results {
				if result.Error != nil {
					fmt.Printf("%s- '%s' -> '%s': %v\n", prefix, result.SourceFile, result.TargetFile, result.Error)
				}
			}
		}
	}
}
