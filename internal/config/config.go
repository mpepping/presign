package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

const (
	DefaultExpiry     = "24h"
	DefaultConfigFile = "~/.config/presign.toml"
)

type Config struct {
	DefaultBucket string `toml:"default_bucket"`
	Expiry        string `toml:"expiry"`
	EndpointURL   string `toml:"endpoint_url"`
	Region        string `toml:"region"`
	Profile       string `toml:"profile"`
	PathStyle     bool   `toml:"path_style"`
}

type FlagOverrides struct {
	Profile     string
	Bucket      string
	EndpointURL string
	Region      string
	Expiry      string
	PathStyle   bool
}

func Load(path string) (*Config, error) {
	if path == "" {
		path = DefaultConfigFile
	}
	path = expandHome(path)

	cfg := &Config{
		Expiry: DefaultExpiry,
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("reading config %s: %w", path, err)
	}

	if err := toml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config %s: %w", path, err)
	}

	if cfg.Expiry == "" {
		cfg.Expiry = DefaultExpiry
	}

	return cfg, nil
}

func (c *Config) ApplyFlags(flags FlagOverrides) {
	if flags.Profile != "" {
		c.Profile = flags.Profile
	}
	if flags.Bucket != "" {
		c.DefaultBucket = flags.Bucket
	}
	if flags.EndpointURL != "" {
		c.EndpointURL = flags.EndpointURL
	}
	if flags.Region != "" {
		c.Region = flags.Region
	}
	if flags.Expiry != "" {
		c.Expiry = flags.Expiry
	}
	if flags.PathStyle {
		c.PathStyle = flags.PathStyle
	}
}

func ParseExpiry(s string) (time.Duration, error) {
	if s == "" {
		s = DefaultExpiry
	}

	// Support "d" suffix for days
	if strings.HasSuffix(s, "d") {
		numStr := strings.TrimSuffix(s, "d")
		days, err := strconv.Atoi(numStr)
		if err != nil {
			return 0, fmt.Errorf("invalid expiry %q: %w", s, err)
		}
		if days <= 0 {
			return 0, fmt.Errorf("invalid expiry %q: must be positive", s)
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}

	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, fmt.Errorf("invalid expiry %q: expected format like '24h' or '7d'", s)
	}
	if d <= 0 {
		return 0, fmt.Errorf("invalid expiry %q: must be positive", s)
	}
	return d, nil
}

func expandHome(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}
