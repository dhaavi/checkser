package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "[dir]",
		Short: "Checkser adds and verifies checksums on folder structures. Alias to check.",
		RunE:  check,
		Args:  cobra.ExactArgs(1),
	}

	checkCmd = &cobra.Command{
		Use:   "check [dir]",
		Short: "Run checkser interactively.",
		RunE:  check,
		Args:  cobra.ExactArgs(1),
	}

	updateCmd = &cobra.Command{
		Use:   "update [dir]",
		Short: "Update checksum files without interaction. Errors are ignored and can be checked using check.",
		RunE:  update,
		Args:  cobra.ExactArgs(1),
	}

	verifyCmd = &cobra.Command{
		Use:   "verify [dir]",
		Short: "Verify checksum files without interaction.",
		RunE:  verify,
		Args:  cobra.ExactArgs(1),
	}

	flagDefaultHash string
	flagRebuild     bool
	flagDigestAll   bool
)

func init() {
	rootCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(verifyCmd)

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

func check(cmd *cobra.Command, args []string) error {
	runInteractive = true
	return run(cmd, args)
}

func update(cmd *cobra.Command, args []string) error {
	runUpdate = true
	return run(cmd, args)
}

func verify(cmd *cobra.Command, args []string) error {
	runVerify = true
	return run(cmd, args)
}
