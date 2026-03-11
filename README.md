# presign

Copy files to S3-compatible object storage and generate presigned URLs for sharing.

## Install

```
go install github.com/mpepping/presign/cmd/presign@latest
```

Or build from source:

```
make build
```

## Usage

```
presign cp <file> [s3://bucket/key]
```

Copies a file to S3 and prints a presigned URL to stdout.

### Quick Start with Config File

Create `~/.config/presign.toml` once:

```toml
default_bucket = "my-shared-files"
endpoint_url = "https://s3.example.com"
region = "us-east-1"
profile = "default"
path_style = true
expiry = "24h"
```

Then upload with minimal commands:

```bash
# Upload using all defaults from config
presign cp report.pdf

# Override specific values via CLI
presign cp report.pdf -e 7d                          # Custom expiry
presign cp report.pdf s3://other-bucket/report.pdf  # Different bucket
presign cp report.pdf --endpoint-url https://other.s3.com  # Override endpoint
```

### Examples

```
# Upload using default bucket from config
presign cp report.pdf

# Upload to a specific bucket and key
presign cp report.pdf s3://my-bucket/docs/report.pdf

# Custom expiry (default: 24h)
presign cp report.pdf -e 7d

# Copy URL to clipboard (macOS)
presign cp report.pdf | pbcopy
```

### Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--bucket` | `-b` | S3 bucket |
| `--expiry` | `-e` | URL expiry duration (e.g. `24h`, `7d`) |
| `--endpoint-url` | | S3 endpoint URL |
| `--region` | | AWS region |
| `--profile` | `-p` | AWS profile |
| `--path-style` | | Use path-style addressing (required for S3-compatible stores like Minio, Hetzner, etc.) |
| `--config` | `-c` | Config file path |
| `--verbose` | `-v` | Verbose output |
| `--content-type` | | Override content type (cp only) |

## Configuration

presign reads AWS credentials from `~/.aws/config` and `~/.aws/credentials`.

An optional config file at `~/.config/presign.toml` can set defaults:

```toml
default_bucket = "my-shared-files"
expiry = "48h"
endpoint_url = "https://minio.example.com"
region = "us-east-1"
profile = "default"
path_style = true
```

**Precedence:** CLI flags > TOML config > AWS config > defaults

## License

MIT
