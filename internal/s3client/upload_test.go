package s3client

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// stubManagerUpload replaces managerUpload for the duration of a test.
// It returns a restore function and a pointer to the captured PutObjectInput.
func stubManagerUpload(t *testing.T, returnErr error) (*s3.PutObjectInput, func()) {
	t.Helper()
	orig := managerUpload
	var captured s3.PutObjectInput
	managerUpload = func(ctx context.Context, client *s3.Client, input *s3.PutObjectInput) error {
		// Drain body so progress callbacks fire.
		if input.Body != nil {
			io.Copy(io.Discard, input.Body)
		}
		captured = *input
		return returnErr
	}
	return &captured, func() { managerUpload = orig }
}

func writeTempFile(t *testing.T, name string, data []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestUploadFileNotFound(t *testing.T) {
	c := &Client{}
	_, err := c.Upload(context.Background(), "bucket", "key", "/nonexistent/file.txt", "", nil)
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	if !strings.Contains(err.Error(), "open") {
		t.Errorf("error = %q, want it to mention open", err)
	}
}

func TestUploadSuccess(t *testing.T) {
	captured, restore := stubManagerUpload(t, nil)
	defer restore()

	path := writeTempFile(t, "test.txt", []byte("hello world"))
	c := &Client{}

	size, err := c.Upload(context.Background(), "my-bucket", "my-key", path, "", nil)
	if err != nil {
		t.Fatalf("Upload() error: %v", err)
	}
	if size != 11 {
		t.Errorf("size = %d, want 11", size)
	}
	if aws.ToString(captured.Bucket) != "my-bucket" {
		t.Errorf("bucket = %q, want %q", aws.ToString(captured.Bucket), "my-bucket")
	}
	if aws.ToString(captured.Key) != "my-key" {
		t.Errorf("key = %q, want %q", aws.ToString(captured.Key), "my-key")
	}
}

func TestUploadExplicitContentType(t *testing.T) {
	captured, restore := stubManagerUpload(t, nil)
	defer restore()

	path := writeTempFile(t, "data.bin", []byte{0x00, 0x01, 0x02})
	c := &Client{}

	_, err := c.Upload(context.Background(), "b", "k", path, "application/custom", nil)
	if err != nil {
		t.Fatalf("Upload() error: %v", err)
	}
	if aws.ToString(captured.ContentType) != "application/custom" {
		t.Errorf("content type = %q, want %q", aws.ToString(captured.ContentType), "application/custom")
	}
}

func TestUploadAutoDetectsContentType(t *testing.T) {
	captured, restore := stubManagerUpload(t, nil)
	defer restore()

	// PNG header bytes trigger image/png detection.
	png := []byte("\x89PNG\r\n\x1a\n" + strings.Repeat("\x00", 100))
	path := writeTempFile(t, "image.png", png)
	c := &Client{}

	_, err := c.Upload(context.Background(), "b", "k", path, "", nil)
	if err != nil {
		t.Fatalf("Upload() error: %v", err)
	}
	ct := aws.ToString(captured.ContentType)
	if ct != "image/png" {
		t.Errorf("auto-detected content type = %q, want %q", ct, "image/png")
	}
}

func TestUploadProgressCallback(t *testing.T) {
	_, restore := stubManagerUpload(t, nil)
	defer restore()

	data := []byte(strings.Repeat("x", 5000))
	path := writeTempFile(t, "big.txt", data)
	c := &Client{}

	var calls int
	var lastRead, lastTotal int64
	progress := func(read, total int64) {
		calls++
		lastRead = read
		lastTotal = total
	}

	size, err := c.Upload(context.Background(), "b", "k", path, "text/plain", progress)
	if err != nil {
		t.Fatalf("Upload() error: %v", err)
	}
	if size != int64(len(data)) {
		t.Errorf("size = %d, want %d", size, len(data))
	}
	if calls == 0 {
		t.Error("progress callback was never called")
	}
	if lastRead != int64(len(data)) {
		t.Errorf("lastRead = %d, want %d", lastRead, len(data))
	}
	if lastTotal != int64(len(data)) {
		t.Errorf("lastTotal = %d, want %d", lastTotal, len(data))
	}
}

func TestUploadNoProgressCallbackWithoutFunc(t *testing.T) {
	captured, restore := stubManagerUpload(t, nil)
	defer restore()

	path := writeTempFile(t, "plain.txt", []byte("data"))
	c := &Client{}

	_, err := c.Upload(context.Background(), "b", "k", path, "", nil)
	if err != nil {
		t.Fatalf("Upload() error: %v", err)
	}
	// Body should be a plain *os.File, not a progressReader.
	if captured.Body == nil {
		t.Fatal("body is nil")
	}
}

func TestUploadS3Error(t *testing.T) {
	_, restore := stubManagerUpload(t, io.ErrUnexpectedEOF)
	defer restore()

	path := writeTempFile(t, "fail.txt", []byte("data"))
	c := &Client{}

	_, err := c.Upload(context.Background(), "b", "k", path, "", nil)
	if err == nil {
		t.Fatal("expected error from S3")
	}
	if !strings.Contains(err.Error(), "upload") {
		t.Errorf("error = %q, want it to mention upload", err)
	}
}

func TestUploadReturnsFileSize(t *testing.T) {
	_, restore := stubManagerUpload(t, nil)
	defer restore()

	data := []byte(strings.Repeat("a", 12345))
	path := writeTempFile(t, "sized.bin", data)
	c := &Client{}

	size, err := c.Upload(context.Background(), "b", "k", path, "application/octet-stream", nil)
	if err != nil {
		t.Fatalf("Upload() error: %v", err)
	}
	if size != 12345 {
		t.Errorf("size = %d, want 12345", size)
	}
}
