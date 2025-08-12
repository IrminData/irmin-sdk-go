package irminutils

import (
	"bytes"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/textproto"
	"path/filepath"
	"strings"
)

// FileContent represents the content and its type for creating multipart files.
// It provides a flexible way to handle different data sources (strings, bytes, readers)
// while maintaining MIME type information and optional size tracking.
type FileContent struct {
	// Data is the source of the file content that implements io.Reader
	Data io.Reader
	// MimeType specifies the MIME type of the content (e.g., "text/plain", "application/json")
	MimeType string
	// Size is the content size in bytes. Set to -1 if unknown or not applicable
	Size int64
}

// FileContentFromString creates FileContent from a string with the specified MIME type.
// This is useful for creating files from text content, JSON strings, or other textual data.
//
// Parameters:
//   - content: The string content to be used as file data
//   - mimeType: The MIME type of the content (e.g., "text/plain", "application/json")
//
// Returns a FileContent instance with the string wrapped in a strings.Reader.
func FileContentFromString(content, mimeType string) *FileContent {
	return &FileContent{
		Data:     strings.NewReader(content),
		MimeType: mimeType,
		Size:     int64(len(content)),
	}
}

// FileContentFromBytes creates FileContent from a byte slice with the specified MIME type.
// This is useful for creating files from binary data, encoded content, or byte arrays.
//
// Parameters:
//   - data: The byte slice containing the file content
//   - mimeType: The MIME type of the content (e.g., "image/png", "application/pdf")
//
// Returns a FileContent instance with the bytes wrapped in a bytes.Reader.
func FileContentFromBytes(data []byte, mimeType string) *FileContent {
	return &FileContent{
		Data:     bytes.NewReader(data),
		MimeType: mimeType,
		Size:     int64(len(data)),
	}
}

// FileContentFromReader creates FileContent from any io.Reader with the specified MIME type and size.
// This is useful for creating files from streams, existing readers, or when you want to
// avoid loading all content into memory.
//
// Parameters:
//   - reader: Any io.Reader that provides the file content
//   - mimeType: The MIME type of the content
//   - size: The size of the content in bytes, or -1 if unknown
//
// Returns a FileContent instance using the provided reader directly.
func FileContentFromReader(reader io.Reader, mimeType string, size int64) *FileContent {
	return &FileContent{
		Data:     reader,
		MimeType: mimeType,
		Size:     size,
	}
}

// AutoDetectMimeType attempts to detect the MIME type from a filename extension.
// It uses Go's built-in mime package to map file extensions to MIME types.
//
// Parameters:
//   - filename: The filename (with extension) to analyze
//
// Returns the detected MIME type, or "application/octet-stream" as a fallback
// if the extension is not recognized.
//
// Examples:
//   - "document.pdf" -> "application/pdf"
//   - "image.png" -> "image/png"
//   - "data.json" -> "application/json"
//   - "unknown.xyz" -> "application/octet-stream"
func AutoDetectMimeType(filename string) string {
	ext := filepath.Ext(filename)
	if mimeType := mime.TypeByExtension(ext); mimeType != "" {
		return mimeType
	}
	return "application/octet-stream" // fallback for unknown extensions
}

// MultipartFormData represents a complete multipart form with the content and metadata
// needed for HTTP multipart/form-data requests.
type MultipartFormData struct {
	// Buffer contains the complete multipart form data
	Buffer *bytes.Buffer
	// ContentType is the full Content-Type header value including the boundary
	ContentType string
	// Boundary is the multipart boundary string used to separate form parts
	Boundary string
}

// CreateMultipartForm creates a complete multipart/form-data payload from FileContent.
// This function generates a properly formatted multipart form that can be used
// directly in HTTP requests for file uploads.
//
// Parameters:
//   - content: The file content to include in the multipart form
//   - filename: The filename to use in the Content-Disposition header
//   - fieldName: The form field name for the file
//
// Returns:
//   - MultipartFormData containing the complete form data and metadata
//   - error if the multipart form creation fails
//
// The generated form includes:
//   - Proper Content-Disposition header with the field name and filename
//   - Content-Type header with the specified MIME type
//   - The file content in the multipart body
//
// Example usage:
//
//	content := FileContentFromString("Hello World", "text/plain")
//	form, err := CreateMultipartForm(content, "hello.txt", "file")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Use form.Buffer and form.ContentType in HTTP request
//	req, _ := http.NewRequest("POST", url, form.Buffer)
//	req.Header.Set("Content-Type", form.ContentType)
func CreateMultipartForm(content *FileContent, filename, fieldName string) (*MultipartFormData, error) {
	// Create a buffer to store the multipart form data
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Create the form file field with proper headers including Content-Type
	headers := make(textproto.MIMEHeader)
	headers.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, fieldName, filename))
	headers.Set("Content-Type", content.MimeType)

	fileWriter, err := writer.CreatePart(headers)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file field: %w", err)
	}

	// Copy the file content to the form field
	_, err = io.Copy(fileWriter, content.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to copy content to form field: %w", err)
	}

	// Close the writer to finalize the multipart form
	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	return &MultipartFormData{
		Buffer:      &buf,
		ContentType: writer.FormDataContentType(),
		Boundary:    writer.Boundary(),
	}, nil
}

// CreateMultipartFormWithFields creates a multipart/form-data payload that includes
// both file content and additional text fields. This is useful when you need to
// submit a file along with metadata or other form data.
//
// Parameters:
//   - content: The file content to include in the multipart form
//   - filename: The filename to use in the Content-Disposition header
//   - fieldName: The form field name for the file
//   - textFields: A map of additional text fields to include in the form
//
// Returns:
//   - MultipartFormData containing the complete form data and metadata
//   - error if the multipart form creation fails
//
// Example usage:
//
//	content := FileContentFromString("Hello World", "text/plain")
//	textFields := map[string]string{
//	    "description": "A sample text file",
//	    "category": "documents",
//	}
//	form, err := CreateMultipartFormWithFields(content, "hello.txt", "file", textFields)
func CreateMultipartFormWithFields(
	content *FileContent,
	filename, fieldName string,
	textFields map[string]string,
) (*MultipartFormData, error) {
	// Create a buffer to store the multipart form data
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add text fields first
	for key, value := range textFields {
		err := writer.WriteField(key, value)
		if err != nil {
			return nil, fmt.Errorf("failed to write text field %q: %w", key, err)
		}
	}

	// Create the form file field with proper headers including Content-Type
	headers := make(textproto.MIMEHeader)
	headers.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, fieldName, filename))
	headers.Set("Content-Type", content.MimeType)

	fileWriter, err := writer.CreatePart(headers)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file field: %w", err)
	}

	// Copy the file content to the form field
	_, err = io.Copy(fileWriter, content.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to copy content to form field: %w", err)
	}

	// Close the writer to finalize the multipart form
	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	return &MultipartFormData{
		Buffer:      &buf,
		ContentType: writer.FormDataContentType(),
		Boundary:    writer.Boundary(),
	}, nil
}
