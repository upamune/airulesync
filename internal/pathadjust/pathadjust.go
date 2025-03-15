package pathadjust

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// PathAdjuster is responsible for adjusting paths in files
type PathAdjuster struct {
	Verbose bool
}

// NewPathAdjuster creates a new path adjuster
func NewPathAdjuster(verbose bool) *PathAdjuster {
	return &PathAdjuster{
		Verbose: verbose,
	}
}

// AdjustmentResult represents the result of a path adjustment operation
type AdjustmentResult struct {
	OriginalPath string
	AdjustedPath string
	LineNumber   int
}

// AdjustPaths adjusts paths in a file based on the relationship between source and target directories
func (p *PathAdjuster) AdjustPaths(sourceFile, targetFile, sourceDir, targetDir string) ([]AdjustmentResult, error) {
	// Read the source file
	content, err := os.ReadFile(sourceFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read source file: %w", err)
	}

	// Detect and adjust paths
	adjustments, adjustedContent, err := p.processContent(content, sourceDir, targetDir)
	if err != nil {
		return nil, fmt.Errorf("failed to process content: %w", err)
	}

	// Ensure the target directory exists
	targetDirPath := filepath.Dir(targetFile)
	if err := os.MkdirAll(targetDirPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create target directory: %w", err)
	}

	// Write the adjusted content to the target file
	if err := os.WriteFile(targetFile, adjustedContent, 0644); err != nil {
		return nil, fmt.Errorf("failed to write target file: %w", err)
	}

	return adjustments, nil
}

// processContent processes the content of a file and adjusts paths
func (p *PathAdjuster) processContent(content []byte, sourceDir, targetDir string) ([]AdjustmentResult, []byte, error) {
	var adjustments []AdjustmentResult
	var outputBuffer bytes.Buffer
	scanner := bufio.NewScanner(bytes.NewReader(content))
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		adjustedLine, lineAdjustments := p.adjustLine(line, lineNum, sourceDir, targetDir)
		adjustments = append(adjustments, lineAdjustments...)
		outputBuffer.WriteString(adjustedLine)
		outputBuffer.WriteString("\n")
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, fmt.Errorf("error scanning content: %w", err)
	}

	return adjustments, outputBuffer.Bytes(), nil
}

// adjustLine adjusts paths in a single line
func (p *PathAdjuster) adjustLine(line string, lineNum int, sourceDir, targetDir string) (string, []AdjustmentResult) {
	var adjustments []AdjustmentResult
	adjustedLine := line

	// Define patterns for path detection
	patterns := []*regexp.Regexp{
		// Import/require statements in various languages
		regexp.MustCompile(`(import|from|require)\s+['"]([./][^'"]+)['"]`),
		// JSON/YAML path references
		regexp.MustCompile(`["'](?:path|file|src|source|location|include)["']\s*:\s*["']([./][^'"]+)["']`),
		// File path references in configuration files
		regexp.MustCompile(`(?:file|path|source|target|output|input)=["']([./][^'"]+)["']`),
		// Markdown links and references
		regexp.MustCompile(`\[.*?\]\(([./][^)]+)\)`),
		// HTML href and src attributes
		regexp.MustCompile(`(?:href|src)=["']([./][^'"]+)["']`),
		// General file paths
		regexp.MustCompile(`["']([./][^'"]+\.(md|txt|json|yaml|yml|js|ts|go|py|java|c|cpp|h|hpp|css|html|xml))["']`),
	}

	for _, pattern := range patterns {
		// Find all matches in the line
		matches := pattern.FindAllStringSubmatchIndex(adjustedLine, -1)

		// Process matches in reverse order to avoid offset issues
		for i := len(matches) - 1; i >= 0; i-- {
			match := matches[i]

			// The path is in the second capturing group (index 2-3)
			// If there's only one capturing group, it's in the first group (index 0-1)
			var pathStartIdx, pathEndIdx int
			if len(match) >= 4 {
				pathStartIdx = match[2]
				pathEndIdx = match[3]
			} else if len(match) >= 2 {
				pathStartIdx = match[0]
				pathEndIdx = match[1]
			} else {
				continue
			}

			originalPath := adjustedLine[pathStartIdx:pathEndIdx]

			// Skip paths that don't start with ./ or ../
			if !strings.HasPrefix(originalPath, "./") && !strings.HasPrefix(originalPath, "../") {
				continue
			}

			// Adjust the path
			adjustedPath, err := p.adjustPath(originalPath, sourceDir, targetDir)
			if err != nil {
				if p.Verbose {
					fmt.Fprintf(os.Stderr, "Warning: Failed to adjust path %s: %v\n", originalPath, err)
				}
				continue
			}

			// Skip if the path didn't change
			if adjustedPath == originalPath {
				continue
			}

			// Replace the path in the line
			adjustedLine = adjustedLine[:pathStartIdx] + adjustedPath + adjustedLine[pathEndIdx:]

			// Record the adjustment
			adjustments = append(adjustments, AdjustmentResult{
				OriginalPath: originalPath,
				AdjustedPath: adjustedPath,
				LineNumber:   lineNum,
			})
		}
	}

	return adjustedLine, adjustments
}

// adjustPath adjusts a single path based on the relationship between source and target directories
func (p *PathAdjuster) adjustPath(path, sourceDir, targetDir string) (string, error) {
	// Convert to absolute paths for calculation
	absSourceDir, err := filepath.Abs(sourceDir)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path for source directory: %w", err)
	}

	absTargetDir, err := filepath.Abs(targetDir)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path for target directory: %w", err)
	}

	// Resolve the original path relative to the source directory
	originalAbsPath := filepath.Join(absSourceDir, path)

	// Calculate the new relative path from the target directory
	newRelPath, err := filepath.Rel(absTargetDir, originalAbsPath)
	if err != nil {
		return "", fmt.Errorf("failed to calculate relative path: %w", err)
	}

	// Ensure the path starts with ./ or ../
	if !strings.HasPrefix(newRelPath, "./") && !strings.HasPrefix(newRelPath, "../") {
		newRelPath = "./" + newRelPath
	}

	return newRelPath, nil
}

// CopyFile copies a file without adjusting paths
func (p *PathAdjuster) CopyFile(sourceFile, targetFile string) error {
	// Open the source file
	src, err := os.Open(sourceFile)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer src.Close()

	// Create the target directory if it doesn't exist
	targetDir := filepath.Dir(targetFile)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Create the target file
	dst, err := os.Create(targetFile)
	if err != nil {
		return fmt.Errorf("failed to create target file: %w", err)
	}
	defer dst.Close()

	// Copy the content
	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	return nil
}

// IsExternalPath checks if a target directory is external to the current repository
func (p *PathAdjuster) IsExternalPath(path string) bool {
	// Check if the path starts with ../ or is an absolute path
	return strings.HasPrefix(path, "../") || filepath.IsAbs(path)
}
