package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mpepping/presign/internal/config"
	"github.com/mpepping/presign/internal/s3client"
	"github.com/mpepping/presign/internal/state"
	"github.com/spf13/cobra"
)

var contentType string

var cpCmd = &cobra.Command{
	Use:   "cp <file>... [s3://bucket/key]",
	Short: "Copy file(s) to S3 and return presigned URLs",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := getConfig(cmd)
		client := getS3Client(cmd)
		ctx := cmd.Context()

		// Determine if the last arg is an S3 destination
		var files []string
		var destBucket, destKey string
		var hasExplicitDest bool

		lastArg := args[len(args)-1]
		if strings.HasPrefix(lastArg, "s3://") {
			var err error
			destBucket, destKey, err = s3client.ParseS3URI(lastArg)
			if err != nil {
				return err
			}
			hasExplicitDest = true
			files = args[:len(args)-1]
			if len(files) == 0 {
				return fmt.Errorf("no files specified")
			}
		} else {
			files = args
		}

		// For a single file with an explicit dest, the key is used as-is (existing behavior).
		// For multiple files with an explicit dest, the key is used as a prefix.
		multiFile := len(files) > 1

		// Parse expiry once
		expiryDuration, err := config.ParseExpiry(cfg.Expiry)
		if err != nil {
			return err
		}

		// Load state once
		statePath, err := state.DefaultPath()
		if err != nil {
			return fmt.Errorf("getting state path: %w", err)
		}
		store, err := state.Load(statePath)
		if err != nil {
			return fmt.Errorf("loading state: %w", err)
		}

		for _, filePath := range files {
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

			// Determine bucket and key for this file
			var bucket, key string
			if hasExplicitDest {
				bucket = destBucket
				if multiFile || destKey == "" || strings.HasSuffix(destKey, "/") {
					key = destKey + filepath.Base(absPath)
				} else {
					key = destKey
				}
			} else {
				bucket = cfg.DefaultBucket
				key = filepath.Base(absPath)
			}

			if bucket == "" {
				return fmt.Errorf("no bucket specified: use --bucket flag or set default_bucket in config")
			}

			if isVerbose() {
				fmt.Fprintf(os.Stderr, "Uploading %s to s3://%s/%s...\n", filePath, bucket, key)
			}

			size, err := client.Upload(ctx, bucket, key, absPath, contentType)
			if err != nil {
				return err
			}

			if isVerbose() {
				fmt.Fprintf(os.Stderr, "Uploaded %d bytes\n", size)
				fmt.Fprintf(os.Stderr, "Generating presigned URL with %s expiry...\n", cfg.Expiry)
			}

			url, expiresAt, err := client.GeneratePresignedURL(ctx, bucket, key, expiryDuration)
			if err != nil {
				return err
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

			if multiFile {
				fmt.Printf("%s:\n", filepath.Base(absPath))
			}
			fmt.Println(url)
		}

		if err := state.Save(statePath, store); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to save state: %v\n", err)
		}

		return nil
	},
}

func init() {
	cpCmd.Flags().StringVar(&contentType, "content-type", "", "override content type")
	rootCmd.AddCommand(cpCmd)
}
