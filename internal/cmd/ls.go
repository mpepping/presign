package cmd

import (
	"fmt"
	"strings"

	"github.com/mpepping/presign/internal/s3client"
	"github.com/spf13/cobra"
)

var lsCmd = &cobra.Command{
	Use:   "ls <s3://bucket[/prefix]>",
	Short: "List objects in an S3 bucket",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := getS3Client(cmd)
		ctx := cmd.Context()

		bucket, prefix, err := s3client.ParseS3URI(args[0])
		if err != nil {
			return err
		}

		objects, prefixes, err := client.ListObjects(ctx, bucket, prefix)
		if err != nil {
			return err
		}

		for _, p := range prefixes {
			fmt.Printf("PRE %s\n", formatKey(p, prefix))
		}
		for _, obj := range objects {
			fmt.Printf("%10d %s\n", obj.Size, formatKey(obj.Key, prefix))
		}

		return nil
	},
}

// formatKey strips the prefix to show relative names, similar to aws s3 ls.
func formatKey(key, prefix string) string {
	return strings.TrimPrefix(key, prefix)
}

func init() {
	rootCmd.AddCommand(lsCmd)
}
