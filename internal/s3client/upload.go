package s3client

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func (c *Client) Upload(ctx context.Context, bucket, key, filePath, contentType string) (int64, error) {
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
		Body:        f,
		ContentType: aws.String(contentType),
	}

	if _, err := c.S3.PutObject(ctx, input); err != nil {
		return 0, fmt.Errorf("upload %s to s3://%s/%s: %w", filePath, bucket, key, err)
	}

	return stat.Size(), nil
}
