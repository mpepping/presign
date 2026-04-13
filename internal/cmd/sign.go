package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/mpepping/presign/internal/config"
	"github.com/mpepping/presign/internal/s3client"
	"github.com/mpepping/presign/internal/state"
	"github.com/spf13/cobra"
)

var signCmd = &cobra.Command{
	Use:   "sign <s3://bucket/key>",
	Short: "Generate a presigned URL for an existing S3 object",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := getConfig(cmd)
		client := getS3Client(cmd)
		ctx := cmd.Context()

		bucket, key, err := s3client.ParseS3URI(args[0])
		if err != nil {
			return err
		}

		if key == "" {
			return fmt.Errorf("no key specified in S3 URI %q", args[0])
		}

		// Verify the object exists
		size, err := client.HeadObject(ctx, bucket, key)
		if err != nil {
			return err
		}

		// Parse expiry
		expiryDuration, err := config.ParseExpiry(cfg.Expiry)
		if err != nil {
			return err
		}

		if isVerbose() {
			fmt.Fprintf(os.Stderr, "Generating presigned URL for s3://%s/%s with %s expiry...\n", bucket, key, cfg.Expiry)
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
			URL:        url,
			ExpiresAt:  expiresAt,
			UploadedAt: time.Now(),
			Size:       size,
		})
		if err := state.Save(statePath, store); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to save state: %v\n", err)
		}

		// Output URL to stdout (pipe-friendly)
		fmt.Println(url)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(signCmd)
}
