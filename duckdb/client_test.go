package duckdb_test

import (
	"fmt"
	"log/slog"
	"testing"

	"github.com/IrminData/irmin-sdk-go/duckdb"
)

func TestNewInMemoryClient(t *testing.T) {
	logger := slog.Default()

	client, err := duckdb.NewInMemoryClient(logger)
	if err != nil {
		t.Fatalf("Failed to create in-memory client: %v", err)
	}
	defer client.Close()

	if client.db == nil {
		t.Error("Database connection should not be nil")
	}
}

func TestCreateTableFromData(t *testing.T) {
	logger := slog.Default()
	client, err := duckdb.NewInMemoryClient(logger)
	if err != nil {
		t.Fatalf("Failed to create in-memory client: %v", err)
	}
	defer client.Close()

	// Test data
	data := []map[string]any{
		{"id": 1, "name": "John", "age": 30, "active": true},
		{"id": 2, "name": "Jane", "age": 25, "active": false},
	}

	err = client.CreateTableFromData("test_users", data)
	if err != nil {
		t.Fatalf("Failed to create table from data: %v", err)
	}

	// Verify the table was created and has data
	rows, err := client.ExecuteQuery("SELECT COUNT(*) FROM test_users")
	if err != nil {
		t.Fatalf("Failed to query table: %v", err)
	}
	if rows.Err() != nil {
		t.Fatalf("Failed to query table: %v", rows.Err())
	}
	defer rows.Close()

	if !rows.Next() {
		t.Fatal("Expected at least one row")
	}

	var count int
	if err := rows.Scan(&count); err != nil {
		t.Fatalf("Failed to scan count: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected 2 rows, got %d", count)
	}
}

func TestQueryToMap(t *testing.T) {
	logger := slog.Default()
	client, err := duckdb.NewInMemoryClient(logger)
	if err != nil {
		t.Fatalf("Failed to create in-memory client: %v", err)
	}
	defer client.Close()

	// Create test data
	data := []map[string]any{
		{"id": 1, "name": "John", "age": 30},
		{"id": 2, "name": "Jane", "age": 25},
	}

	err = client.CreateTableFromData("test_users", data)
	if err != nil {
		t.Fatalf("Failed to create table from data: %v", err)
	}

	// Query and convert to map
	results, err := client.QueryToMap("SELECT * FROM test_users ORDER BY id")
	if err != nil {
		t.Fatalf("Failed to query to map: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	// Check first row
	first := results[0]
	if first["name"] != "John" {
		t.Errorf("Expected name 'John', got %v", first["name"])
	}
}

func TestMergeDataSources(t *testing.T) {
	logger := slog.Default()
	client, err := duckdb.NewInMemoryClient(logger)
	if err != nil {
		t.Fatalf("Failed to create in-memory client: %v", err)
	}
	defer client.Close()

	// Create test data sources
	dataSources := map[string][]map[string]any{
		"source1": {
			{"id": 1, "name": "John", "age": 30},
			{"id": 2, "name": "Jane", "age": 25},
		},
		"source2": {
			{"id": 3, "name": "Bob", "age": 35},
			{"id": 4, "name": "Alice", "age": 28},
		},
	}

	result, err := client.MergeDataSources(dataSources, "merged_users", duckdb.MergeStrategyUnion)
	if err != nil {
		t.Fatalf("Failed to merge data sources: %v", err)
	}

	if result.TableName != "merged_users" {
		t.Errorf("Expected table name 'merged_users', got %s", result.TableName)
	}

	if result.RowCount != 4 {
		t.Errorf("Expected 4 rows, got %d", result.RowCount)
	}

	if len(result.SourceNames) != 2 {
		t.Errorf("Expected 2 source names, got %d", len(result.SourceNames))
	}

	// Verify the merged data
	rows, err := client.QueryToMap("SELECT COUNT(*) as count FROM merged_users")
	if err != nil {
		t.Fatalf("Failed to query merged table: %v", err)
	}

	if len(rows) != 1 {
		t.Errorf("Expected 1 row in count query, got %d", len(rows))
	}
}

func TestGetDuckDBReadOptions(t *testing.T) {
	tests := []struct {
		filename    string
		expectError bool
		readFunc    string
	}{
		{"data.csv", false, "read_csv_auto"},
		{"data.json", false, "read_json_auto"},
		{"data.parquet", false, "read_parquet"},
		{"data.jsonl", false, "read_json_auto"},
		{"data.unknown", true, ""},
	}

	for _, test := range tests {
		t.Run(test.filename, func(t *testing.T) {
			options, err := duckdb.GetDuckDBReadOptions(test.filename)

			if test.expectError {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", test.filename)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error for %s: %v", test.filename, err)
				return
			}

			if options.ReadFunction != test.readFunc {
				t.Errorf("Expected read function %s for %s, got %s",
					test.readFunc, test.filename, options.ReadFunction)
			}
		})
	}
}

func TestIsFormatSupported(t *testing.T) {
	supportedFormats := []string{
		"data.csv", "data.json", "data.parquet", "data.jsonl",
		"data.tsv", "data.avro", "data.orc",
	}

	unsupportedFormats := []string{
		"data.txt", "data.unknown", "data.exe",
	}

	for _, format := range supportedFormats {
		if !duckdb.IsFormatSupported(format) {
			t.Errorf("Format %s should be supported", format)
		}
	}

	for _, format := range unsupportedFormats {
		if duckdb.IsFormatSupported(format) {
			t.Errorf("Format %s should not be supported", format)
		}
	}
}

func TestLoadFileFromBytes(t *testing.T) {
	logger := slog.Default()
	client, err := duckdb.NewInMemoryClient(logger)
	if err != nil {
		t.Fatalf("Failed to create in-memory client: %v", err)
	}
	defer client.Close()

	// Test CSV data
	csvData := []byte(`name,age,city
John,30,New York
Jane,25,Los Angeles
Bob,35,Chicago`)

	err = client.LoadFileFromBytes(csvData, "users.csv", "users_from_csv")
	if err != nil {
		t.Fatalf("Failed to load CSV from bytes: %v", err)
	}

	// Verify the data was loaded correctly
	results, err := client.QueryToMap("SELECT COUNT(*) as count FROM users_from_csv")
	if err != nil {
		t.Fatalf("Failed to query loaded table: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 row in count query, got %d", len(results))
	}

	count, ok := results[0]["count"]
	if !ok {
		t.Error("Expected count column in result")
	}

	// DuckDB might return different numeric types, so convert to string for comparison
	countStr := fmt.Sprintf("%v", count)
	if countStr != "3" {
		t.Errorf("Expected count of 3, got %s", countStr)
	}

	// Test JSON data
	jsonData := []byte(`[
		{"name": "Alice", "age": 28, "city": "Boston"},
		{"name": "Charlie", "age": 32, "city": "Seattle"}
	]`)

	err = client.LoadFileFromBytes(jsonData, "users.json", "users_from_json")
	if err != nil {
		t.Fatalf("Failed to load JSON from bytes: %v", err)
	}

	// Verify JSON data
	jsonResults, err := client.QueryToMap("SELECT name FROM users_from_json ORDER BY name")
	if err != nil {
		t.Fatalf("Failed to query JSON table: %v", err)
	}

	if len(jsonResults) != 2 {
		t.Errorf("Expected 2 rows from JSON, got %d", len(jsonResults))
	}

	if jsonResults[0]["name"] != "Alice" {
		t.Errorf("Expected first name to be Alice, got %v", jsonResults[0]["name"])
	}
}
