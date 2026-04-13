package s3client

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type ListObject struct {
	Key  string
	Size int64
}

// ListObjects lists objects in a bucket with the given prefix.
func (c *Client) ListObjects(ctx context.Context, bucket, prefix string) ([]ListObject, []string, error) {
	var objects []ListObject
	var prefixes []string

	paginator := s3.NewListObjectsV2Paginator(c.S3, &s3.ListObjectsV2Input{
		Bucket:    aws.String(bucket),
		Prefix:    aws.String(prefix),
		Delimiter: aws.String("/"),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("listing s3://%s/%s: %w", bucket, prefix, err)
		}

		for _, p := range page.CommonPrefixes {
			prefixes = append(prefixes, aws.ToString(p.Prefix))
		}

		for _, obj := range page.Contents {
			objects = append(objects, listObjectFromS3(obj))
		}
	}

	return objects, prefixes, nil
}

func listObjectFromS3(obj types.Object) ListObject {
	return ListObject{
		Key:  aws.ToString(obj.Key),
		Size: aws.ToInt64(obj.Size),
	}
}
