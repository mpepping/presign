package s3client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// managerUpload performs the S3 upload via the transfer manager.
// Replaced in tests to avoid real S3 calls.
var managerUpload = func(ctx context.Context, client *s3.Client, input *s3.PutObjectInput) error {
	uploader := manager.NewUploader(client)
	_, err := uploader.Upload(ctx, input)
	return err
}

func (c *Client) Upload(ctx context.Context, bucket, key, filePath, contentType string, onProgress ProgressFunc) (int64, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return 0, fmt.Errorf("open %s: %w", filePath, err)
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return 0, fmt.Errorf("stat %s: %w", filePath, err)
	}

	if contentType == "" {
		buf := make([]byte, 512)
		n, _ := f.Read(buf)
		contentType = http.DetectContentType(buf[:n])
		if _, err := f.Seek(0, 0); err != nil {
			return 0, fmt.Errorf("seek %s: %w", filePath, err)
		}
	}

	input := &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}

	var body io.Reader = f
	if onProgress != nil {
		body = newProgressReader(f, stat.Size(), onProgress)
	}
	input.Body = body

	if err := managerUpload(ctx, c.S3, input); err != nil {
		return 0, fmt.Errorf("upload %s to s3://%s/%s: %w", filePath, bucket, key, err)
	}

	return stat.Size(), nil
}
