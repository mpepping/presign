package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestParseExpiry(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    time.Duration
		wantErr bool
	}{
		{name: "empty defaults to 24h", input: "", want: 24 * time.Hour},
		{name: "hours", input: "12h", want: 12 * time.Hour},
		{name: "minutes", input: "30m", want: 30 * time.Minute},
		{name: "days", input: "7d", want: 7 * 24 * time.Hour},
		{name: "one day", input: "1d", want: 24 * time.Hour},
		{name: "negative hours", input: "-1h", wantErr: true},
		{name: "zero days", input: "0d", wantErr: true},
		{name: "negative days", input: "-3d", wantErr: true},
		{name: "invalid format", input: "abc", wantErr: true},
		{name: "invalid day format", input: "xd", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseExpiry(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseExpiry(%q) expected error, got %v", tt.input, got)
				}
				return
			}
			if err != nil {
				t.Errorf("ParseExpiry(%q) unexpected error: %v", tt.input, err)
				return
			}
			if got != tt.want {
				t.Errorf("ParseExpiry(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestLoadDefaults(t *testing.T) {
	cfg, err := Load("/nonexistent/path/presign.toml")
	if err != nil {
		t.Fatalf("Load() with missing file should not error, got: %v", err)
	}
	if cfg.Expiry != DefaultExpiry {
		t.Errorf("Expiry = %q, want %q", cfg.Expiry, DefaultExpiry)
	}
	if cfg.DefaultBucket != "" {
		t.Errorf("DefaultBucket = %q, want empty", cfg.DefaultBucket)
	}
}

func TestLoadFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "presign.toml")

	content := `default_bucket = "my-bucket"
expiry = "48h"
region = "eu-west-1"
endpoint_url = "https://s3.example.com"
path_style = true
profile = "myprofile"
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.DefaultBucket != "my-bucket" {
		t.Errorf("DefaultBucket = %q, want %q", cfg.DefaultBucket, "my-bucket")
	}
	if cfg.Expiry != "48h" {
		t.Errorf("Expiry = %q, want %q", cfg.Expiry, "48h")
	}
	if cfg.Region != "eu-west-1" {
		t.Errorf("Region = %q, want %q", cfg.Region, "eu-west-1")
	}
	if cfg.EndpointURL != "https://s3.example.com" {
		t.Errorf("EndpointURL = %q, want %q", cfg.EndpointURL, "https://s3.example.com")
	}
	if !cfg.PathStyle {
		t.Error("PathStyle = false, want true")
	}
	if cfg.Profile != "myprofile" {
		t.Errorf("Profile = %q, want %q", cfg.Profile, "myprofile")
	}
}

func TestLoadEmptyExpiry(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "presign.toml")

	content := `default_bucket = "test"
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.Expiry != DefaultExpiry {
		t.Errorf("Expiry = %q, want default %q", cfg.Expiry, DefaultExpiry)
	}
}

func TestLoadInvalidTOML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.toml")
	if err := os.WriteFile(path, []byte("{{invalid}}"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Error("Load() with invalid TOML should return error")
	}
}

func TestApplyFlags(t *testing.T) {
	cfg := &Config{
		DefaultBucket: "original",
		Expiry:        "24h",
		Region:        "us-east-1",
	}

	cfg.ApplyFlags(FlagOverrides{
		Bucket: "override-bucket",
		Region: "eu-west-1",
		Expiry: "7d",
	})

	if cfg.DefaultBucket != "override-bucket" {
		t.Errorf("DefaultBucket = %q, want %q", cfg.DefaultBucket, "override-bucket")
	}
	if cfg.Region != "eu-west-1" {
		t.Errorf("Region = %q, want %q", cfg.Region, "eu-west-1")
	}
	if cfg.Expiry != "7d" {
		t.Errorf("Expiry = %q, want %q", cfg.Expiry, "7d")
	}
}

func TestApplyFlagsEmpty(t *testing.T) {
	cfg := &Config{
		DefaultBucket: "keep",
		Profile:       "keep-profile",
		EndpointURL:   "https://keep.example.com",
		Region:        "us-west-2",
		Expiry:        "12h",
	}

	cfg.ApplyFlags(FlagOverrides{})

	if cfg.DefaultBucket != "keep" {
		t.Errorf("DefaultBucket changed to %q", cfg.DefaultBucket)
	}
	if cfg.Profile != "keep-profile" {
		t.Errorf("Profile changed to %q", cfg.Profile)
	}
	if cfg.EndpointURL != "https://keep.example.com" {
		t.Errorf("EndpointURL changed to %q", cfg.EndpointURL)
	}
	if cfg.Region != "us-west-2" {
		t.Errorf("Region changed to %q", cfg.Region)
	}
	if cfg.Expiry != "12h" {
		t.Errorf("Expiry changed to %q", cfg.Expiry)
	}
}

func TestExpandHome(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home directory")
	}

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "tilde path", input: "~/foo/bar", want: filepath.Join(home, "foo/bar")},
		{name: "absolute path", input: "/tmp/test", want: "/tmp/test"},
		{name: "relative path", input: "foo/bar", want: "foo/bar"},
		{name: "just tilde slash", input: "~/", want: home},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := expandHome(tt.input)
			if got != tt.want {
				t.Errorf("expandHome(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
