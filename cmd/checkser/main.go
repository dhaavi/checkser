package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "checkser",
		Short: "Checkser adds and verifies checksums on folder structures.",
		Run:   run,
	}

	flagDefaultHash string
	flagRebuild     bool
	flagDigestAll   bool
)

func init() {
	rootCmd.PersistentFlags().StringVar(&flagDefaultHash, "default-hash", "", "define default hash algorithm to be used")
	rootCmd.PersistentFlags().BoolVar(&flagRebuild, "rebuild", false, "complete rebuild: all files are digested, all checksum files rewritten (produces virtual changes)")
	rootCmd.PersistentFlags().BoolVar(&flagDigestAll, "digest-all", false, "always digest files, not only when size/modtime changed")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) {
	// Do Stuff Here
}
