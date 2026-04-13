package s3client

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// HeadObject checks that an object exists and returns its size.
func (c *Client) HeadObject(ctx context.Context, bucket, key string) (int64, error) {
	resp, err := c.S3.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return 0, fmt.Errorf("object s3://%s/%s not found: %w", bucket, key, err)
	}

	var size int64
	if resp.ContentLength != nil {
		size = *resp.ContentLength
	}
	return size, nil
}
