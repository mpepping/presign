package s3client

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func (c *Client) GeneratePresignedURL(ctx context.Context, bucket, key string, expiry time.Duration) (string, time.Time, error) {
	presignClient := s3.NewPresignClient(c.S3)

	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	resp, err := presignClient.PresignGetObject(ctx, input, s3.WithPresignExpires(expiry))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("presign s3://%s/%s: %w", bucket, key, err)
	}

	expiresAt := time.Now().Add(expiry)
	return resp.URL, expiresAt, nil
}
