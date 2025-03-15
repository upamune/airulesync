package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/upamune/airulesync/internal/config"
)

// FileInfo represents information about a file to be synchronized
type FileInfo struct {
	SourcePath      string
	SourceDir       string
	RelativePath    string
	Pattern         string
	AdjustPaths     bool
	Overwrite       bool
	SourceDirConfig *config.SourceDir
}

// Scanner is responsible for scanning directories for files to synchronize
type Scanner struct {
	Config *config.Config
}

// NewScanner creates a new scanner
func NewScanner(cfg *config.Config) *Scanner {
	return &Scanner{
		Config: cfg,
	}
}

// ScanSourceDirs scans all source directories for files to synchronize
func (s *Scanner) ScanSourceDirs() ([]FileInfo, error) {
	var files []FileInfo

	for _, sourceDir := range s.Config.SourceDirs {
		dirFiles, err := s.scanSourceDir(sourceDir)
		if err != nil {
			return nil, fmt.Errorf("failed to scan source directory %s: %w", sourceDir.Path, err)
		}
		files = append(files, dirFiles...)
	}

	return files, nil
}

// scanSourceDir scans a single source directory for files to synchronize
func (s *Scanner) scanSourceDir(sourceDir config.SourceDir) ([]FileInfo, error) {
	var files []FileInfo
	dirOverwrite := sourceDir.GetDirectoryOverwrite()

	for _, fileSpec := range sourceDir.Files {
		pattern := fileSpec.GetPattern()
		adjustPaths := fileSpec.ShouldAdjustPaths()
		overwrite := fileSpec.ShouldOverwrite(dirOverwrite)

		// Check if the pattern is a glob pattern
		if strings.ContainsAny(pattern, "*?[") {
			// Handle glob pattern
			matches, err := s.findGlobMatches(sourceDir.Path, pattern, sourceDir.IgnoreFiles)
			if err != nil {
				return nil, fmt.Errorf("failed to find glob matches for pattern %s: %w", pattern, err)
			}

			for _, match := range matches {
				relPath, err := filepath.Rel(sourceDir.Path, match)
				if err != nil {
					return nil, fmt.Errorf("failed to get relative path for %s: %w", match, err)
				}

				files = append(files, FileInfo{
					SourcePath:      match,
					SourceDir:       sourceDir.Path,
					RelativePath:    relPath,
					Pattern:         pattern,
					AdjustPaths:     adjustPaths,
					Overwrite:       overwrite,
					SourceDirConfig: &sourceDir,
				})
			}
		} else {
			// Handle simple file pattern
			fullPath := filepath.Join(sourceDir.Path, pattern)
			if s.shouldIgnoreFile(fullPath, sourceDir.IgnoreFiles) {
				continue
			}

			// Check if the file exists
			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				// Skip non-existent files
				continue
			} else if err != nil {
				return nil, fmt.Errorf("failed to stat file %s: %w", fullPath, err)
			}

			files = append(files, FileInfo{
				SourcePath:      fullPath,
				SourceDir:       sourceDir.Path,
				RelativePath:    pattern,
				Pattern:         pattern,
				AdjustPaths:     adjustPaths,
				Overwrite:       overwrite,
				SourceDirConfig: &sourceDir,
			})
		}
	}

	return files, nil
}

// findGlobMatches finds all files matching a glob pattern
func (s *Scanner) findGlobMatches(basePath, pattern string, ignorePatterns []string) ([]string, error) {
	fullPattern := filepath.Join(basePath, pattern)
	matches, err := filepath.Glob(fullPattern)
	if err != nil {
		return nil, fmt.Errorf("failed to glob pattern %s: %w", fullPattern, err)
	}

	// Filter out ignored files
	var filteredMatches []string
	for _, match := range matches {
		if !s.shouldIgnoreFile(match, ignorePatterns) {
			// Check if it's a file (not a directory)
			info, err := os.Stat(match)
			if err != nil {
				return nil, fmt.Errorf("failed to stat file %s: %w", match, err)
			}

			if !info.IsDir() {
				filteredMatches = append(filteredMatches, match)
			}
		}
	}

	return filteredMatches, nil
}

// shouldIgnoreFile checks if a file should be ignored
func (s *Scanner) shouldIgnoreFile(filePath string, ignorePatterns []string) bool {
	for _, ignorePattern := range ignorePatterns {
		// Check if the ignore pattern is a glob pattern
		if strings.ContainsAny(ignorePattern, "*?[") {
			matches, err := filepath.Match(ignorePattern, filepath.Base(filePath))
			if err == nil && matches {
				return true
			}

			// Try matching against the full path
			fullIgnorePattern := filepath.Join(filepath.Dir(filePath), ignorePattern)
			matches, err = filepath.Match(fullIgnorePattern, filePath)
			if err == nil && matches {
				return true
			}
		} else {
			// Simple pattern matching
			if filepath.Base(filePath) == ignorePattern {
				return true
			}

			// Check if the full path matches
			if filePath == ignorePattern {
				return true
			}
		}
	}

	return false
}

// ScanDirectory scans a directory for rule files (used by the init command)
func (s *Scanner) ScanDirectory(dir string) ([]string, error) {
	var ruleFiles []string

	// Common rule file patterns
	patterns := []string{
		".clinerules",
		".cursor/rules/*.mdc",
		".roomodes",
		".rooignore",
		".cursorignore",
	}

	for _, pattern := range patterns {
		fullPattern := filepath.Join(dir, pattern)
		matches, err := filepath.Glob(fullPattern)
		if err != nil {
			return nil, fmt.Errorf("failed to glob pattern %s: %w", fullPattern, err)
		}

		for _, match := range matches {
			// Check if it's a file (not a directory)
			info, err := os.Stat(match)
			if err != nil {
				return nil, fmt.Errorf("failed to stat file %s: %w", match, err)
			}

			if !info.IsDir() {
				relPath, err := filepath.Rel(dir, match)
				if err != nil {
					return nil, fmt.Errorf("failed to get relative path for %s: %w", match, err)
				}
				ruleFiles = append(ruleFiles, relPath)
			}
		}
	}

	return ruleFiles, nil
}

// FindPotentialTargetDirs finds potential target directories for rule files
func (s *Scanner) FindPotentialTargetDirs(baseDir string) ([]string, error) {
	var targetDirs []string

	err := filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the base directory itself
		if path == baseDir {
			return nil
		}

		// Only consider directories
		if !info.IsDir() {
			return nil
		}

		// Skip hidden directories (except .cursor)
		if strings.HasPrefix(filepath.Base(path), ".") && filepath.Base(path) != ".cursor" {
			return filepath.SkipDir
		}

		// Skip vendor directories
		if filepath.Base(path) == "vendor" || filepath.Base(path) == "node_modules" {
			return filepath.SkipDir
		}

		// Check if this is a potential target directory
		// We're looking for directories that:
		// 1. Are not the base directory
		// 2. Contain source code files (Go, JavaScript, TypeScript, etc.)
		// 3. Don't already have rule files

		// Check for source code files
		hasSourceFiles := false
		entries, err := os.ReadDir(path)
		if err != nil {
			return err
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			ext := strings.ToLower(filepath.Ext(entry.Name()))
			if ext == ".go" || ext == ".js" || ext == ".ts" || ext == ".py" || ext == ".java" || ext == ".c" || ext == ".cpp" {
				hasSourceFiles = true
				break
			}
		}

		if hasSourceFiles {
			// Check if it already has rule files
			hasRuleFiles := false
			rulePatterns := []string{".clinerules", ".cursor/rules", ".roomodes", ".rooignore", ".cursorignore"}
			for _, pattern := range rulePatterns {
				if _, err := os.Stat(filepath.Join(path, pattern)); err == nil {
					hasRuleFiles = true
					break
				}
			}

			if !hasRuleFiles {
				relPath, err := filepath.Rel(baseDir, path)
				if err != nil {
					return err
				}
				targetDirs = append(targetDirs, relPath)
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory %s: %w", baseDir, err)
	}

	return targetDirs, nil
}
