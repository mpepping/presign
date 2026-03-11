package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/mpepping/presign/internal/config"
	"github.com/mpepping/presign/internal/s3client"
	"github.com/spf13/cobra"
)

type contextKey string

const (
	configKey   contextKey = "config"
	s3ClientKey contextKey = "s3client"
)

var (
	cfgFile     string
	profile     string
	bucket      string
	endpointURL string
	region      string
	expiry      string
	pathStyle   bool
	verbose     bool
)

var rootCmd = &cobra.Command{
	Use:   "presign",
	Short: "Copy files to S3 and generate presigned URLs",
	Long:  "presign copies files to S3-compatible object storage and generates presigned URLs for sharing.",
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		cfg.ApplyFlags(config.FlagOverrides{
			Profile:     profile,
			Bucket:      bucket,
			EndpointURL: endpointURL,
			Region:      region,
			Expiry:      expiry,
			PathStyle:   pathStyle,
		})

		ctx := context.WithValue(cmd.Context(), configKey, cfg)

		client, err := s3client.NewClient(ctx, cfg)
		if err != nil {
			return fmt.Errorf("creating S3 client: %w", err)
		}
		ctx = context.WithValue(ctx, s3ClientKey, client)
		cmd.SetContext(ctx)

		if verbose {
			fmt.Fprintf(os.Stderr, "Config: bucket=%s region=%s endpoint=%s path_style=%v expiry=%s\n",
				cfg.DefaultBucket, cfg.Region, cfg.EndpointURL, cfg.PathStyle, cfg.Expiry)
		}

		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default ~/.config/presign.toml)")
	rootCmd.PersistentFlags().StringVarP(&profile, "profile", "p", "", "AWS profile")
	rootCmd.PersistentFlags().StringVarP(&bucket, "bucket", "b", "", "S3 bucket")
	rootCmd.PersistentFlags().StringVar(&endpointURL, "endpoint-url", "", "S3 endpoint URL")
	rootCmd.PersistentFlags().StringVar(&region, "region", "", "AWS region")
	rootCmd.PersistentFlags().StringVarP(&expiry, "expiry", "e", "", "presigned URL expiry (e.g. 24h, 7d)")
	rootCmd.PersistentFlags().BoolVar(&pathStyle, "path-style", false, "use path-style addressing (required for most S3-compatible stores)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
}

func Execute() error {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
	return err
}

func getConfig(cmd *cobra.Command) *config.Config {
	return cmd.Context().Value(configKey).(*config.Config)
}

func getS3Client(cmd *cobra.Command) *s3client.Client {
	return cmd.Context().Value(s3ClientKey).(*s3client.Client)
}

func isVerbose() bool {
	return verbose
}
