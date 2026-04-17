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
presign cp <file>... [s3://bucket/key]
presign sign <s3://bucket/key>
presign ls <s3://bucket[/prefix]>
```

`cp` copies one or more files to S3 and prints a presigned URL per file to stdout.
`sign` generates a presigned URL for an object already in S3.
`ls` lists objects in an S3 bucket.

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

# Upload multiple files at once
presign cp *.pdf

# Override specific values via CLI
presign cp report.pdf -e 7d                          # Custom expiry
presign cp report.pdf s3://other-bucket/report.pdf  # Different bucket
presign cp report.pdf --endpoint-url https://other.s3.com  # Override endpoint
```

### Examples

Upload and sign using the `cp` subcommand:


```
# Upload using default bucket from config
presign cp report.pdf

# Upload to a specific bucket and key
presign cp report.pdf s3://my-bucket/docs/report.pdf

# Custom expiry (default: 24h)
presign cp report.pdf -e 7d

# Copy URL to clipboard (macOS)
presign cp report.pdf | pbcopy

# Upload with custom endpoint and expiry
presign cp presentation.pptx --endpoint-url https://s3.backblazeb2.com -e 14d

# Upload with path-style addressing (for Minio, Hetzner, etc.)
presign cp backup.tar.gz --path-style

# Upload with specific AWS profile
presign cp data.csv -p production

# Upload multiple files to default bucket
presign cp *.csv

# Upload multiple files to a specific prefix
presign cp report.pdf slides.pptx s3://my-bucket/meeting/
```


Sign or resign files already in the remote using the `sign` subcommand:


```
# Generate a presigned URL for an existing object
presign sign s3://my-bucket/docs/report.pdf

# With custom expiry
presign sign s3://my-bucket/docs/report.pdf -e 7d

# Copy presigned URL to clipboard (macOS)
presign sign s3://my-bucket/docs/report.pdf | pbcopy

# With specific AWS profile and endpoint
presign sign s3://my-bucket/archive.tar.gz -p production --endpoint-url https://s3.example.com
```

List bucket contents using the `ls` subcommand:


```
# List objects under a prefix
presign ls s3://my-bucket/docs/

# List top-level contents of a bucket
presign ls s3://my-bucket/

# With specific AWS profile and endpoint
presign ls s3://my-bucket/docs/ -p production --endpoint-url https://s3.example.com
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
