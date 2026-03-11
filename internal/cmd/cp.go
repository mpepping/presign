package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mpepping/presign/internal/config"
	"github.com/mpepping/presign/internal/s3client"
	"github.com/mpepping/presign/internal/state"
	"github.com/spf13/cobra"
)

var contentType string

var cpCmd = &cobra.Command{
	Use:   "cp <file> [s3://bucket/key]",
	Short: "Copy a file to S3 and return a presigned URL",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]
		cfg := getConfig(cmd)
		client := getS3Client(cmd)
		ctx := cmd.Context()

		// Validate local file
		absPath, err := filepath.Abs(filePath)
		if err != nil {
			return fmt.Errorf("resolving path %s: %w", filePath, err)
		}
		info, err := os.Stat(absPath)
		if err != nil {
			return fmt.Errorf("file %s: %w", filePath, err)
		}
		if info.IsDir() {
			return fmt.Errorf("%s is a directory, use 'presign sync' instead", filePath)
		}

		// Determine bucket and key
		var bucket, key string
		if len(args) == 2 {
			bucket, key, err = s3client.ParseS3URI(args[1])
			if err != nil {
				return err
			}
		} else {
			bucket = cfg.DefaultBucket
			key = filepath.Base(absPath)
		}

		if bucket == "" {
			return fmt.Errorf("no bucket specified: use --bucket flag or set default_bucket in config")
		}
		if key == "" {
			key = filepath.Base(absPath)
		}

		// Parse expiry
		expiryDuration, err := config.ParseExpiry(cfg.Expiry)
		if err != nil {
			return err
		}

		if isVerbose() {
			fmt.Fprintf(os.Stderr, "Uploading %s to s3://%s/%s...\n", filePath, bucket, key)
		}

		// Upload
		size, err := client.Upload(ctx, bucket, key, absPath, contentType)
		if err != nil {
			return err
		}

		if isVerbose() {
			fmt.Fprintf(os.Stderr, "Uploaded %d bytes\n", size)
			fmt.Fprintf(os.Stderr, "Generating presigned URL with %s expiry...\n", cfg.Expiry)
		}

		// Generate presigned URL
		url, expiresAt, err := client.GeneratePresignedURL(ctx, bucket, key, expiryDuration)
		if err != nil {
			return err
		}

		// Save state
		statePath, err := state.DefaultPath()
		if err != nil {
			return fmt.Errorf("getting state path: %w", err)
		}
		store, err := state.Load(statePath)
		if err != nil {
			return fmt.Errorf("loading state: %w", err)
		}
		store.Add(state.Entry{
			Bucket:     bucket,
			Key:        key,
			LocalPath:  absPath,
			URL:        url,
			ExpiresAt:  expiresAt,
			UploadedAt: time.Now(),
			Size:       size,
		})
		if err := state.Save(statePath, store); err != nil {
			// Non-fatal: warn but still output the URL
			fmt.Fprintf(os.Stderr, "Warning: failed to save state: %v\n", err)
		}

		// Output URL to stdout (pipe-friendly)
		fmt.Println(url)
		return nil
	},
}

func init() {
	cpCmd.Flags().StringVar(&contentType, "content-type", "", "override content type")
	rootCmd.AddCommand(cpCmd)
}
