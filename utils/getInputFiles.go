package irminutils

import (
	"fmt"
	"os"
	"path/filepath"
)

// GetInputFile reads a file from the _input directory.
// Returns the file content as bytes and any error encountered.
func GetInputFile(filePath string) ([]byte, error) {
	// Construct the full path to the input file
	fullPath := filepath.Join("_input", filePath)

	// Read the file content
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read input file %s: %w", filePath, err)
	}

	return content, nil
}
