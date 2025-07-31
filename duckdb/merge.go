package duckdb

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// MergeStrategy defines how to handle conflicts when merging data.
type MergeStrategy string

const (
	// MergeStrategyUnion combines all rows from all sources (allows duplicates).
	MergeStrategyUnion MergeStrategy = "union"

	// MergeStrategyUnionDistinct combines all rows but removes exact duplicates.
	MergeStrategyUnionDistinct MergeStrategy = "union_distinct"

	// MergeStrategyFirstWins keeps only rows from the first source when conflicts occur.
	MergeStrategyFirstWins MergeStrategy = "first_wins"

	// MergeStrategyLastWins keeps only rows from the last source when conflicts occur.
	MergeStrategyLastWins MergeStrategy = "last_wins"

	// decrementStep is used when iterating backwards through source table names.
	decrementStep = 2
)

// MergeResult represents the result of merging multiple data sources.
type MergeResult struct {
	TableName   string   `json:"table_name"`
	RowCount    int      `json:"row_count"`
	SourceNames []string `json:"source_names"` // Track which sources contributed
}

// MergeDataSources merges multiple in-memory data sources into a single table.
// It handles different merge strategies for conflict resolution.
//
// Parameters:
//   - dataSources: Map of source names to their data
//   - targetTableName: The name of the target table to create
//   - strategy: How to handle merge conflicts
//
// Returns the merge result or an error.
func (c *InMemoryClient) MergeDataSources(
	dataSources map[string][]map[string]any,
	targetTableName string,
	strategy MergeStrategy,
) (*MergeResult, error) {
	if len(dataSources) == 0 {
		return nil, errors.New("no data sources provided for merging")
	}

	// If only one source, create table directly
	if len(dataSources) == 1 {
		for sourceName, data := range dataSources {
			if err := c.CreateTableFromData(targetTableName, data); err != nil {
				return nil, fmt.Errorf("failed to create table from single source: %w", err)
			}
			return &MergeResult{
				TableName:   targetTableName,
				RowCount:    len(data),
				SourceNames: []string{sourceName},
			}, nil
		}
	}

	// Create individual tables for each source
	var tempTableNames []string
	var sourceNames []string
	totalRows := 0

	for sourceName, data := range dataSources {
		tempTableName := fmt.Sprintf("temp_%s_%s", targetTableName, cleanTableName(sourceName))
		tempTableNames = append(tempTableNames, tempTableName)
		sourceNames = append(sourceNames, sourceName)
		totalRows += len(data)

		if err := c.CreateTableFromData(tempTableName, data); err != nil {
			return nil, fmt.Errorf("failed to create temp table for source %s: %w", sourceName, err)
		}
	}

	// Build merge query based on strategy
	mergeQuery, err := c.buildMergeQuery(tempTableNames, targetTableName, strategy)
	if err != nil {
		return nil, fmt.Errorf("failed to build merge query: %w", err)
	}

	// Execute merge query
	if _, execErr := c.db.Exec(mergeQuery); execErr != nil {
		return nil, fmt.Errorf("failed to execute merge query: %w", execErr)
	}

	// Get actual row count from merged table
	var finalRowCount int
	safeTargetTableName, validationErr := validateSQLIdentifierForMerge(targetTableName)
	if validationErr != nil {
		c.logger.Warn("invalid target table name for count query", "table", targetTableName, "error", validationErr)
		finalRowCount = totalRows // fallback
	} else {
		countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s", safeTargetTableName)
		if scanErr := c.db.QueryRow(countQuery).Scan(&finalRowCount); scanErr != nil {
			finalRowCount = totalRows // fallback
		}
	}

	// Clean up temporary tables
	for _, tempTable := range tempTableNames {
		dropQuery := fmt.Sprintf("DROP TABLE IF EXISTS %s", tempTable)
		if _, dropErr := c.db.Exec(dropQuery); dropErr != nil {
			c.logger.Warn("failed to drop temporary table", "table", tempTable, "error", dropErr)
		}
	}

	return &MergeResult{
		TableName:   targetTableName,
		RowCount:    finalRowCount,
		SourceNames: sourceNames,
	}, nil
}

// MergeFiles merges multiple files from byte content into a single table.
// This is useful for processing files loaded into memory.
func (c *InMemoryClient) MergeFiles(
	sourceFiles map[string][]byte,
	targetTableName string,
	strategy MergeStrategy,
) (*MergeResult, error) {
	if len(sourceFiles) == 0 {
		return nil, errors.New("no source files provided for merging")
	}

	// Create temporary files and load them as tables
	var tempTableNames []string
	var sourceNames []string
	var cleanup []func()

	defer func() {
		for _, cleanupFunc := range cleanup {
			cleanupFunc()
		}
	}()

	for filename, content := range sourceFiles {
		// Create temporary file
		tempFile, err := os.CreateTemp("", fmt.Sprintf("duckdb_merge_*_%s", filename))
		if err != nil {
			return nil, fmt.Errorf("failed to create temp file for %s: %w", filename, err)
		}

		tempFilePath := tempFile.Name()
		cleanup = append(cleanup, func() {
			if removeErr := os.Remove(tempFilePath); removeErr != nil {
				// Log error but don't fail the operation since this is cleanup
			}
		})

		// Write content to temp file
		if _, writeErr := tempFile.Write(content); writeErr != nil {
			if closeErr := tempFile.Close(); closeErr != nil {
				// Log both errors but prioritize the write error
			}
			return nil, fmt.Errorf("failed to write content to temp file: %w", writeErr)
		}
		if closeErr := tempFile.Close(); closeErr != nil {
			return nil, fmt.Errorf("failed to close temp file: %w", closeErr)
		}

		// Create table from file
		tempTableName := fmt.Sprintf("temp_%s_%s", targetTableName, cleanTableName(filename))
		tempTableNames = append(tempTableNames, tempTableName)
		sourceNames = append(sourceNames, filename)

		if err := c.loadFileAsTableFromPath(tempFilePath, filename, tempTableName); err != nil {
			return nil, fmt.Errorf("failed to load file %s as table: %w", filename, err)
		}
	}

	// Build and execute merge query
	mergeQuery, err := c.buildMergeQuery(tempTableNames, targetTableName, strategy)
	if err != nil {
		return nil, fmt.Errorf("failed to build merge query: %w", err)
	}

	if _, execErr := c.db.Exec(mergeQuery); execErr != nil {
		return nil, fmt.Errorf("failed to execute merge query: %w", execErr)
	}

	// Get row count
	var finalRowCount int
	safeTargetTableName, validationErr := validateSQLIdentifierForMerge(targetTableName)
	if validationErr != nil {
		c.logger.Warn("invalid target table name for count query", "table", targetTableName, "error", validationErr)
	} else {
		countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s", safeTargetTableName)
		if scanErr := c.db.QueryRow(countQuery).Scan(&finalRowCount); scanErr != nil {
			c.logger.Warn("failed to get row count", "error", scanErr)
		}
	}

	// Clean up temporary tables
	for _, tempTable := range tempTableNames {
		dropQuery := fmt.Sprintf("DROP TABLE IF EXISTS %s", tempTable)
		if _, dropErr := c.db.Exec(dropQuery); dropErr != nil {
			c.logger.Warn("failed to drop temporary table", "table", tempTable, "error", dropErr)
		}
	}

	return &MergeResult{
		TableName:   targetTableName,
		RowCount:    finalRowCount,
		SourceNames: sourceNames,
	}, nil
}

// loadFileAsTable overloaded version that accepts byte data.
func (c *InMemoryClient) loadFileAsTable(data []byte, originalFilename, tableName string) error {
	// Validate format is supported before proceeding
	if !IsFormatSupported(originalFilename) {
		return fmt.Errorf("unsupported format for %s", originalFilename)
	}

	// Create temporary file from byte data
	tempFile, err := os.CreateTemp("", fmt.Sprintf("duckdb_load_*_%s", originalFilename))
	if err != nil {
		return fmt.Errorf("failed to create temp file for %s: %w", originalFilename, err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Write byte data to temp file
	if _, writeErr := tempFile.Write(data); writeErr != nil {
		return fmt.Errorf("failed to write data to temp file: %w", writeErr)
	}
	if closeErr := tempFile.Close(); closeErr != nil {
		return fmt.Errorf("failed to close temp file before reading: %w", closeErr)
	}

	// Create table from file using the file path version
	return c.loadFileAsTableFromPath(tempFile.Name(), originalFilename, tableName)
}

// loadFileAsTableFromPath loads a file from a file path into DuckDB as a table.
func (c *InMemoryClient) loadFileAsTableFromPath(filePath, originalFilename, tableName string) error {
	options, err := GetDuckDBReadOptions(originalFilename)
	if err != nil {
		return fmt.Errorf("unsupported format for %s: %w", originalFilename, err)
	}

	// Install required extensions
	for _, ext := range GetRequiredExtensions(options) {
		installQuery := fmt.Sprintf("INSTALL %s;", ext)
		loadQuery := fmt.Sprintf("LOAD %s;", ext)

		if _, installErr := c.db.Exec(installQuery); installErr != nil {
			c.logger.Warn("failed to install extension", "extension", ext, "error", installErr)
		}
		if _, loadErr := c.db.Exec(loadQuery); loadErr != nil {
			c.logger.Warn("failed to load extension", "extension", ext, "error", loadErr)
		}
	}

	// Create table from file
	readQuery := BuildReadQuery(filePath, options)
	safeTableName, err := validateSQLIdentifierForMerge(tableName)
	if err != nil {
		return fmt.Errorf("invalid table name: %w", err)
	}
	createQuery := fmt.Sprintf("CREATE TABLE %s AS SELECT * FROM %s", safeTableName, readQuery)

	if _, createErr := c.db.Exec(createQuery); createErr != nil {
		return fmt.Errorf("failed to create table %s from file: %w", tableName, createErr)
	}

	return nil
}

// buildMergeQuery constructs the appropriate merge query based on strategy.
func (c *InMemoryClient) buildMergeQuery(
	sourceTableNames []string,
	targetTableName string,
	strategy MergeStrategy,
) (string, error) {
	if len(sourceTableNames) == 0 {
		return "", errors.New("no source tables provided")
	}

	var selectQueries []string
	for _, tableName := range sourceTableNames {
		selectQueries = append(selectQueries, fmt.Sprintf("SELECT * FROM %s", tableName))
	}

	switch strategy {
	case MergeStrategyUnion:
		query := fmt.Sprintf("CREATE TABLE %s AS (%s)",
			targetTableName,
			strings.Join(selectQueries, " UNION ALL "))
		return query, nil

	case MergeStrategyUnionDistinct:
		query := fmt.Sprintf("CREATE TABLE %s AS (%s)",
			targetTableName,
			strings.Join(selectQueries, " UNION "))
		return query, nil

	case MergeStrategyFirstWins:
		// For first wins, we take the first table and use EXCEPT to remove duplicates from others
		if len(sourceTableNames) == 1 {
			return fmt.Sprintf("CREATE TABLE %s AS SELECT * FROM %s", targetTableName, sourceTableNames[0]), nil
		}

		baseQuery := fmt.Sprintf("SELECT * FROM %s", sourceTableNames[0])
		for i := 1; i < len(sourceTableNames); i++ {
			baseQuery = fmt.Sprintf("(%s) UNION (SELECT * FROM %s EXCEPT %s)",
				baseQuery, sourceTableNames[i], baseQuery)
		}
		return fmt.Sprintf("CREATE TABLE %s AS %s", targetTableName, baseQuery), nil

	case MergeStrategyLastWins:
		// For last wins, we reverse the order and apply first wins logic
		if len(sourceTableNames) == 1 {
			return fmt.Sprintf("CREATE TABLE %s AS SELECT * FROM %s", targetTableName, sourceTableNames[0]), nil
		}

		baseQuery := fmt.Sprintf("SELECT * FROM %s", sourceTableNames[len(sourceTableNames)-1])
		for i := len(sourceTableNames) - decrementStep; i >= 0; i-- {
			baseQuery = fmt.Sprintf("(%s) UNION (SELECT * FROM %s EXCEPT %s)",
				baseQuery, sourceTableNames[i], baseQuery)
		}
		return fmt.Sprintf("CREATE TABLE %s AS %s", targetTableName, baseQuery), nil

	default:
		return "", fmt.Errorf("unsupported merge strategy: %s", strategy)
	}
}

// cleanTableName removes special characters from table names to make them valid SQL identifiers.
func cleanTableName(name string) string {
	// Replace common problematic characters
	name = strings.ReplaceAll(name, ".", "_")
	name = strings.ReplaceAll(name, "-", "_")
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")

	// Remove file extensions
	if idx := strings.LastIndex(name, "."); idx != -1 {
		name = name[:idx]
	}

	return name
}

// validateSQLIdentifierForMerge validates and safely quotes SQL identifiers for merge operations.
// This helps prevent SQL injection by ensuring only valid identifiers are used.
func validateSQLIdentifierForMerge(identifier string) (string, error) {
	// Check for valid SQL identifier (alphanumeric and underscore only)
	validIdentifier := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
	if !validIdentifier.MatchString(identifier) {
		return "", fmt.Errorf("invalid SQL identifier: %s", identifier)
	}
	// Return quoted identifier to prevent SQL injection
	return fmt.Sprintf(`"%s"`, identifier), nil
}
