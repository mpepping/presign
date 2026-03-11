package s3client

import (
	"testing"
)

func TestParseS3URI(t *testing.T) {
	tests := []struct {
		name       string
		uri        string
		wantBucket string
		wantKey    string
		wantErr    bool
	}{
		{
			name:       "full URI",
			uri:        "s3://my-bucket/path/to/file.txt",
			wantBucket: "my-bucket",
			wantKey:    "path/to/file.txt",
		},
		{
			name:       "bucket only",
			uri:        "s3://my-bucket",
			wantBucket: "my-bucket",
			wantKey:    "",
		},
		{
			name:       "bucket with trailing slash",
			uri:        "s3://my-bucket/",
			wantBucket: "my-bucket",
			wantKey:    "",
		},
		{
			name:       "simple key",
			uri:        "s3://bucket/file.txt",
			wantBucket: "bucket",
			wantKey:    "file.txt",
		},
		{
			name:       "nested key",
			uri:        "s3://bucket/a/b/c/d.txt",
			wantBucket: "bucket",
			wantKey:    "a/b/c/d.txt",
		},
		{
			name:    "missing s3 prefix",
			uri:     "https://bucket/key",
			wantErr: true,
		},
		{
			name:    "empty string",
			uri:     "",
			wantErr: true,
		},
		{
			name:    "just s3 prefix",
			uri:     "s3://",
			wantErr: true,
		},
		{
			name:    "no bucket after prefix",
			uri:     "s3:///key",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bucket, key, err := ParseS3URI(tt.uri)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseS3URI(%q) expected error, got bucket=%q key=%q", tt.uri, bucket, key)
				}
				return
			}
			if err != nil {
				t.Errorf("ParseS3URI(%q) unexpected error: %v", tt.uri, err)
				return
			}
			if bucket != tt.wantBucket {
				t.Errorf("ParseS3URI(%q) bucket = %q, want %q", tt.uri, bucket, tt.wantBucket)
			}
			if key != tt.wantKey {
				t.Errorf("ParseS3URI(%q) key = %q, want %q", tt.uri, key, tt.wantKey)
			}
		})
	}
}
