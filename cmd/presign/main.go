package main

import (
	"os"

	"github.com/mpepping/presign/internal/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
