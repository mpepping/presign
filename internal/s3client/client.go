package s3client

import (
	"context"
	"fmt"
	"strings"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/mpepping/presign/internal/config"
)

type Client struct {
	S3 *s3.Client
}

func NewClient(ctx context.Context, cfg *config.Config) (*Client, error) {
	var opts []func(*awsconfig.LoadOptions) error

	if cfg.Profile != "" {
		opts = append(opts, awsconfig.WithSharedConfigProfile(cfg.Profile))
	}
	if cfg.Region != "" {
		opts = append(opts, awsconfig.WithRegion(cfg.Region))
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("loading AWS config: %w", err)
	}

	s3Client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if cfg.EndpointURL != "" {
			o.BaseEndpoint = &cfg.EndpointURL
		}
		if cfg.PathStyle {
			o.UsePathStyle = true
		}
	})

	return &Client{S3: s3Client}, nil
}

// ParseS3URI parses an S3 URI like "s3://bucket/key/path" into bucket and key.
func ParseS3URI(uri string) (bucket, key string, err error) {
	if !strings.HasPrefix(uri, "s3://") {
		return "", "", fmt.Errorf("invalid S3 URI %q: expected format s3://bucket/key", uri)
	}

	rest := strings.TrimPrefix(uri, "s3://")
	if rest == "" {
		return "", "", fmt.Errorf("invalid S3 URI %q: missing bucket", uri)
	}

	parts := strings.SplitN(rest, "/", 2)
	bucket = parts[0]
	if bucket == "" {
		return "", "", fmt.Errorf("invalid S3 URI %q: empty bucket", uri)
	}

	if len(parts) == 2 {
		key = parts[1]
	}

	return bucket, key, nil
}
