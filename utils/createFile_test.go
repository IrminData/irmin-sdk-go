package irminutils_test

import (
	"bytes"
	"io"
	"mime/multipart"
	"strings"
	"testing"

	irminutils "github.com/IrminData/irmin-sdk-go/utils"
)

// TestFileContentFromString tests creating FileContent from string input
func TestFileContentFromString(t *testing.T) {
	tests := []struct {
		name         string
		content      string
		mimeType     string
		expectedSize int64
	}{
		{
			name:         "simple text content",
			content:      "Hello World",
			mimeType:     "text/plain",
			expectedSize: 11,
		},
		{
			name:         "json content",
			content:      `{"key": "value"}`,
			mimeType:     "application/json",
			expectedSize: 16,
		},
		{
			name:         "empty content",
			content:      "",
			mimeType:     "text/plain",
			expectedSize: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileContent := irminutils.FileContentFromString(tt.content, tt.mimeType)

			// Check MIME type
			if fileContent.MimeType != tt.mimeType {
				t.Errorf("expected MIME type %q, got %q", tt.mimeType, fileContent.MimeType)
			}

			// Check size
			expectedSize := int64(len(tt.content))
			if fileContent.Size != expectedSize {
				t.Errorf("expected size %d, got %d", expectedSize, fileContent.Size)
			}

			// Check that we can read the content
			data, err := io.ReadAll(fileContent.Data)
			if err != nil {
				t.Fatalf("failed to read content: %v", err)
			}

			if string(data) != tt.content {
				t.Errorf("expected content %q, got %q", tt.content, string(data))
			}
		})
	}
}

// TestFileContentFromBytes tests creating FileContent from byte slice input
func TestFileContentFromBytes(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		mimeType string
	}{
		{
			name:     "text bytes",
			data:     []byte("Hello World"),
			mimeType: "text/plain",
		},
		{
			name:     "binary data",
			data:     []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, // PNG header
			mimeType: "image/png",
		},
		{
			name:     "empty bytes",
			data:     []byte{},
			mimeType: "application/octet-stream",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileContent := irminutils.FileContentFromBytes(tt.data, tt.mimeType)

			// Check MIME type
			if fileContent.MimeType != tt.mimeType {
				t.Errorf("expected MIME type %q, got %q", tt.mimeType, fileContent.MimeType)
			}

			// Check size
			if fileContent.Size != int64(len(tt.data)) {
				t.Errorf("expected size %d, got %d", len(tt.data), fileContent.Size)
			}

			// Check that we can read the content
			data, err := io.ReadAll(fileContent.Data)
			if err != nil {
				t.Fatalf("failed to read content: %v", err)
			}

			if !bytes.Equal(data, tt.data) {
				t.Errorf("expected data %v, got %v", tt.data, data)
			}
		})
	}
}

// TestFileContentFromReader tests creating FileContent from io.Reader input
func TestFileContentFromReader(t *testing.T) {
	content := "Hello from reader"
	mimeType := "text/plain"
	size := int64(17)

	reader := strings.NewReader(content)
	fileContent := irminutils.FileContentFromReader(reader, mimeType, size)

	// Check MIME type
	if fileContent.MimeType != mimeType {
		t.Errorf("expected MIME type %q, got %q", mimeType, fileContent.MimeType)
	}

	// Check size
	if fileContent.Size != size {
		t.Errorf("expected size %d, got %d", size, fileContent.Size)
	}

	// Check that we can read the content
	data, err := io.ReadAll(fileContent.Data)
	if err != nil {
		t.Fatalf("failed to read content: %v", err)
	}

	if string(data) != content {
		t.Errorf("expected content %q, got %q", content, string(data))
	}
}

// TestAutoDetectMimeType tests MIME type detection from file extensions
func TestAutoDetectMimeType(t *testing.T) {
	tests := []struct {
		filename       string
		expectedPrefix string // We'll check if the result starts with this
	}{
		{"document.pdf", "application/pdf"},
		{"image.png", "image/png"},
		{"image.jpg", "image/jpeg"},
		{"data.json", "application/json"},
		{"style.css", "text/css"},
		{"script.js", "text/javascript"},
		{"page.html", "text/html"},
		{"data.xml", "application/xml"}, // Could be text/xml or application/xml
		{"archive.zip", "application/zip"},
		{"document.txt", "text/plain"},
		{"no_extension", "application/octet-stream"}, // fallback
		{"", "application/octet-stream"},             // fallback
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := irminutils.AutoDetectMimeType(tt.filename)
			if !strings.HasPrefix(result, tt.expectedPrefix) {
				t.Errorf("expected MIME type to start with %q for %q, got %q", tt.expectedPrefix, tt.filename, result)
			}
		})
	}

	// Test specific case for unknown extension - it might map to something unexpected
	t.Run("unknown.xyz", func(t *testing.T) {
		result := irminutils.AutoDetectMimeType("unknown.xyz")
		// The system might know about .xyz files, so we just ensure we get something
		if result == "" {
			t.Error("expected non-empty MIME type for unknown.xyz")
		}
		// If it doesn't know, it should fall back to application/octet-stream
		if !strings.Contains(result, "/") {
			t.Errorf("expected valid MIME type format for unknown.xyz, got %q", result)
		}
	})
}

// TestCreateMultipartForm tests basic multipart form creation
func TestCreateMultipartForm(t *testing.T) {
	content := "Hello World"
	mimeType := "text/plain"
	filename := "hello.txt"
	fieldName := "file"

	fileContent := irminutils.FileContentFromString(content, mimeType)
	form, err := irminutils.CreateMultipartForm(fileContent, filename, fieldName)
	if err != nil {
		t.Fatalf("failed to create multipart form: %v", err)
	}

	// Validate form metadata
	if form.Buffer == nil {
		t.Error("expected non-nil buffer")
	}
	if form.ContentType == "" {
		t.Error("expected non-empty content type")
	}
	if form.Boundary == "" {
		t.Error("expected non-empty boundary")
	}
	if !strings.Contains(form.ContentType, "multipart/form-data") {
		t.Errorf("expected content type to contain 'multipart/form-data', got %q", form.ContentType)
	}
	if !strings.Contains(form.ContentType, form.Boundary) {
		t.Errorf("expected content type to contain boundary %q, got %q", form.Boundary, form.ContentType)
	}

	// Parse and validate form content
	reader := multipart.NewReader(form.Buffer, form.Boundary)
	part, partErr := reader.NextPart()
	if partErr != nil {
		t.Fatalf("failed to read multipart part: %v", partErr)
	}

	if part.FormName() != fieldName {
		t.Errorf("expected form name %q, got %q", fieldName, part.FormName())
	}
	if part.FileName() != filename {
		t.Errorf("expected filename %q, got %q", filename, part.FileName())
	}

	partContent, readErr := io.ReadAll(part)
	if readErr != nil {
		t.Fatalf("failed to read part content: %v", readErr)
	}
	if string(partContent) != content {
		t.Errorf("expected content %q, got %q", content, string(partContent))
	}

	// Ensure there are no more parts
	_, nextPartErr := reader.NextPart()
	if nextPartErr != io.EOF {
		t.Error("expected only one part in multipart form")
	}
}

// TestCreateMultipartFormWithFields tests multipart form creation with additional text fields
func TestCreateMultipartFormWithFields(t *testing.T) {
	content := "Hello World"
	mimeType := "text/plain"
	filename := "hello.txt"
	fieldName := "file"
	textFields := map[string]string{
		"description": "A sample text file",
		"category":    "documents",
	}

	fileContent := irminutils.FileContentFromString(content, mimeType)
	form, err := irminutils.CreateMultipartFormWithFields(fileContent, filename, fieldName, textFields)
	if err != nil {
		t.Fatalf("failed to create multipart form with fields: %v", err)
	}

	// Parse the multipart form to verify all fields
	reader := multipart.NewReader(form.Buffer, form.Boundary)

	foundTextFields := make(map[string]string)
	foundFile := false

	for {
		part, partErr := reader.NextPart()
		if partErr == io.EOF {
			break
		}
		if partErr != nil {
			t.Fatalf("failed to read multipart part: %v", partErr)
		}

		partContent, readErr := io.ReadAll(part)
		if readErr != nil {
			t.Fatalf("failed to read part content: %v", readErr)
		}

		if part.FileName() != "" {
			// This is the file part
			foundFile = true
			validateFilePart(t, part, filename, fieldName, content, partContent)
		} else {
			// This is a text field
			foundTextFields[part.FormName()] = string(partContent)
		}
	}

	// Verify we found the file
	if !foundFile {
		t.Error("expected to find file part in multipart form")
	}

	// Verify all text fields were included
	validateTextFields(t, foundTextFields, textFields)
}

// validateFilePart is a helper function to validate file part content
func validateFilePart(
	t *testing.T,
	part *multipart.Part,
	expectedFilename, expectedFieldName, expectedContent string,
	actualContent []byte,
) {
	t.Helper()

	if part.FormName() != expectedFieldName {
		t.Errorf("expected file form name %q, got %q", expectedFieldName, part.FormName())
	}
	if part.FileName() != expectedFilename {
		t.Errorf("expected filename %q, got %q", expectedFilename, part.FileName())
	}
	if string(actualContent) != expectedContent {
		t.Errorf("expected file content %q, got %q", expectedContent, string(actualContent))
	}
}

// validateTextFields is a helper function to validate text field content
func validateTextFields(t *testing.T, found, expected map[string]string) {
	t.Helper()

	for key, expectedValue := range expected {
		if foundValue, exists := found[key]; !exists {
			t.Errorf("expected to find text field %q", key)
		} else if foundValue != expectedValue {
			t.Errorf("expected text field %q to have value %q, got %q", key, expectedValue, foundValue)
		}
	}

	if len(found) != len(expected) {
		t.Errorf("expected %d text fields, found %d", len(expected), len(found))
	}
}

// TestCreateMultipartFormEmptyContent tests handling of empty content
func TestCreateMultipartFormEmptyContent(t *testing.T) {
	fileContent := irminutils.FileContentFromString("", "text/plain")
	form, err := irminutils.CreateMultipartForm(fileContent, "empty.txt", "file")
	if err != nil {
		t.Fatalf("failed to create multipart form with empty content: %v", err)
	}

	// Parse the form to verify it handles empty content correctly
	reader := multipart.NewReader(form.Buffer, form.Boundary)
	part, partErr := reader.NextPart()
	if partErr != nil {
		t.Fatalf("failed to read multipart part: %v", partErr)
	}

	content, readErr := io.ReadAll(part)
	if readErr != nil {
		t.Fatalf("failed to read part content: %v", readErr)
	}

	if len(content) != 0 {
		t.Errorf("expected empty content, got %d bytes", len(content))
	}
}

// TestCreateMultipartFormLargeContent tests handling of larger content
func TestCreateMultipartFormLargeContent(t *testing.T) {
	// Create a larger content string
	largeContent := strings.Repeat("Hello World! ", 1000) // ~13KB
	fileContent := irminutils.FileContentFromString(largeContent, "text/plain")
	form, err := irminutils.CreateMultipartForm(fileContent, "large.txt", "file")
	if err != nil {
		t.Fatalf("failed to create multipart form with large content: %v", err)
	}

	// Parse the form to verify content integrity
	reader := multipart.NewReader(form.Buffer, form.Boundary)
	part, partErr := reader.NextPart()
	if partErr != nil {
		t.Fatalf("failed to read multipart part: %v", partErr)
	}

	content, readErr := io.ReadAll(part)
	if readErr != nil {
		t.Fatalf("failed to read part content: %v", readErr)
	}

	if string(content) != largeContent {
		t.Error("large content was not preserved correctly")
	}
}

// TestCreateMultipartFormContentType tests that the Content-Type header is properly set in multipart form file parts
func TestCreateMultipartFormContentType(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		mimeType  string
		filename  string
		fieldName string
	}{
		{
			name:      "text/plain content",
			content:   "Hello World",
			mimeType:  "text/plain",
			filename:  "hello.txt",
			fieldName: "file",
		},
		{
			name:      "application/json content",
			content:   `{"key": "value"}`,
			mimeType:  "application/json",
			filename:  "data.json",
			fieldName: "document",
		},
		{
			name:      "image/png content",
			content:   "fake png content",
			mimeType:  "image/png",
			filename:  "image.png",
			fieldName: "upload",
		},
		{
			name:      "application/pdf content",
			content:   "fake pdf content",
			mimeType:  "application/pdf",
			filename:  "document.pdf",
			fieldName: "attachment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileContent := irminutils.FileContentFromString(tt.content, tt.mimeType)
			form, err := irminutils.CreateMultipartForm(fileContent, tt.filename, tt.fieldName)
			if err != nil {
				t.Fatalf("failed to create multipart form: %v", err)
			}

			// Parse the multipart form to check the Content-Type header
			reader := multipart.NewReader(form.Buffer, form.Boundary)
			part, partErr := reader.NextPart()
			if partErr != nil {
				t.Fatalf("failed to read multipart part: %v", partErr)
			}

			// Check the Content-Type header of the file part
			contentType := part.Header.Get("Content-Type")
			if contentType != tt.mimeType {
				t.Errorf("expected Content-Type %q, got %q", tt.mimeType, contentType)
			}

			// Verify basic part properties are still correct
			if part.FormName() != tt.fieldName {
				t.Errorf("expected form name %q, got %q", tt.fieldName, part.FormName())
			}
			if part.FileName() != tt.filename {
				t.Errorf("expected filename %q, got %q", tt.filename, part.FileName())
			}
		})
	}
}

// TestCreateMultipartFormWithFieldsContentType tests that Content-Type header is properly set when using additional fields
func TestCreateMultipartFormWithFieldsContentType(t *testing.T) {
	content := "Hello World"
	mimeType := "text/plain"
	filename := "hello.txt"
	fieldName := "file"
	textFields := map[string]string{
		"description": "A sample text file",
	}

	fileContent := irminutils.FileContentFromString(content, mimeType)
	form, err := irminutils.CreateMultipartFormWithFields(fileContent, filename, fieldName, textFields)
	if err != nil {
		t.Fatalf("failed to create multipart form with fields: %v", err)
	}

	// Parse the multipart form to find the file part and check its Content-Type
	reader := multipart.NewReader(form.Buffer, form.Boundary)

	for {
		part, partErr := reader.NextPart()
		if partErr == io.EOF {
			break
		}
		if partErr != nil {
			t.Fatalf("failed to read multipart part: %v", partErr)
		}

		// Check if this is the file part (has a filename)
		if part.FileName() != "" {
			contentType := part.Header.Get("Content-Type")
			if contentType != mimeType {
				t.Errorf("expected Content-Type %q, got %q", mimeType, contentType)
			}
			return // Found and validated the file part
		}
	}

	t.Error("file part not found in multipart form")
}

// TestMultipartFormDataStruct tests the MultipartFormData struct properties
func TestMultipartFormDataStruct(t *testing.T) {
	fileContent := irminutils.FileContentFromString("test", "text/plain")
	form, err := irminutils.CreateMultipartForm(fileContent, "test.txt", "file")
	if err != nil {
		t.Fatalf("failed to create multipart form: %v", err)
	}

	// Test that all fields are properly set
	if form.Buffer == nil {
		t.Error("Buffer should not be nil")
	}
	if form.ContentType == "" {
		t.Error("ContentType should not be empty")
	}
	if form.Boundary == "" {
		t.Error("Boundary should not be empty")
	}
	if !strings.Contains(form.ContentType, form.Boundary) {
		t.Error("ContentType should contain the boundary")
	}

	// Test that we can read from buffer multiple times by creating a copy
	originalLen := form.Buffer.Len()
	bufferCopy := bytes.NewBuffer(form.Buffer.Bytes())

	if bufferCopy.Len() != originalLen {
		t.Error("Buffer copy should have same length as original")
	}
}
