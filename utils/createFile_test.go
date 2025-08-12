package irminutils_test

import (
	"io"
	"strings"
	"testing"

	irminutils "github.com/IrminData/irmin-sdk-go/utils"
)

func TestNewFile(t *testing.T) {
	content := "Hello World"
	filename := "test.txt"

	file := irminutils.NewFile(content, filename)
	if string(file.Content) != content {
		t.Errorf("Expected content %q, got %q", content, string(file.Content))
	}
	if file.Filename != filename {
		t.Errorf("Expected filename %q, got %q", filename, file.Filename)
	}
}

func TestFileReader(t *testing.T) {
	content := "Hello World"
	file := irminutils.NewFile(content, "test.txt")
	reader := file.Reader()

	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("Failed to read: %v", err)
	}
	if string(data) != content {
		t.Errorf("Expected %q, got %q", content, string(data))
	}
}

func TestFileMultipartFile(t *testing.T) {
	content := "Hello World"
	filename := "test.txt"

	file := irminutils.NewFile(content, filename)
	mFile := file.MultipartFile()
	defer mFile.Close()

	// Test reading
	data, err := io.ReadAll(mFile)
	if err != nil {
		t.Fatalf("Failed to read: %v", err)
	}
	if string(data) != content {
		t.Errorf("Expected %q, got %q", content, string(data))
	}

	// Test seeking
	pos, err := mFile.Seek(6, io.SeekStart)
	if err != nil {
		t.Fatalf("Failed to seek: %v", err)
	}
	if pos != 6 {
		t.Errorf("Expected position 6, got %d", pos)
	}
}

func TestCreateMultipartForm(t *testing.T) {
	file := irminutils.NewFile("Hello World", "test.txt")
	buf, contentType, err := irminutils.CreateMultipartForm(file, "file")
	if err != nil {
		t.Fatalf("Failed to create form: %v", err)
	}

	if buf.Len() == 0 {
		t.Error("Expected non-empty buffer")
	}
	if !strings.Contains(contentType, "multipart/form-data") {
		t.Errorf("Expected multipart content type, got %q", contentType)
	}
}

func TestCreateMultipartFormWithFields(t *testing.T) {
	file := irminutils.NewFile("Hello World", "test.txt")
	textFields := map[string]string{"description": "Test file"}

	buf, contentType, err := irminutils.CreateMultipartFormWithFields(file, "file", textFields)
	if err != nil {
		t.Fatalf("Failed to create form: %v", err)
	}

	if buf.Len() == 0 {
		t.Error("Expected non-empty buffer")
	}
	if !strings.Contains(contentType, "multipart/form-data") {
		t.Errorf("Expected multipart content type, got %q", contentType)
	}
}
