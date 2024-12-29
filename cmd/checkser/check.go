package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dhaavi/checkser"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(checkCmd)
}

var checkCmd = &cobra.Command{
	Use:  "check [dir]",
	RunE: check,
	Args: cobra.ExactArgs(1),
}

func check(cmd *cobra.Command, args []string) error {
	dir, err := filepath.Abs(args[0])
	if err != nil {
		return fmt.Errorf("invalid directory: %w", err)
	}

	// Scan the directory for checksums and files.
	fmt.Println("Finding files and directories...")
	scan, err := checkser.ScanDir(dir, false)
	if err != nil {
		return fmt.Errorf("invalid directory: %w", err)
	}
	for _, line := range scan.FmtFindStatus() {
		fmt.Println(line)
	}
	fmt.Println("")

	// Prompt before continuing when there are errors.
	cliReader := bufio.NewReader(os.Stdin)
	if scan.Stats.FindingErrors.Load() > 0 {
	actionFind:
		for {
			if lessIsAvailable() {
				fmt.Printf("Encountered %d errors during scan - continue? [Y]es, [q]uit, [v]iew errors (with less): ", scan.Stats.FindingErrors.Load())
			} else {
				fmt.Printf("Encountered %d errors during scan - continue? [Y]es, [q]uit, [v]iew errors: ", scan.Stats.FindingErrors.Load())
			}
			line, err := cliReader.ReadString('\n')
			if err != nil {
				fmt.Printf("failed to read action: %s\n", err)
			}
			switch strings.TrimSpace(line) {
			case "Y", "y", "":
				break actionFind

			case "Q", "q":
				return nil

			case "V", "v":
				viewDetails(scan, checkser.Failed)
			}
		}
	}

	// Digest any needed files.
	fmt.Println("Digesting files...")
	scan.DigestFiles(true)
	for _, line := range scan.FmtDigestStatus() {
		fmt.Println(line)
	}
	fmt.Println("")

	// Calculate changes.
	fmt.Println("Detected Changes:")
	scan.CalculateChangeStats()
	for _, line := range scan.FmtChangeStatus() {
		fmt.Println(line)
	}
	fmt.Println("")

action:
	for {
		if lessIsAvailable() {
			fmt.Printf("Apply? [Y]es, [q]uit, [v]iew changes (with less): [a]dded, [r]emoved, [c]hanged, [n]o change, [f]ailed: ")
		} else {
			fmt.Printf("Apply? [Y]es, [q]uit, [v]iew changes: [a]dded, [r]emoved, [c]hanged, [n]o change, [f]ailed: ")
		}
		line, err := cliReader.ReadString('\n')
		if err != nil {
			fmt.Printf("failed to read action: %s\n", err)
		}
		switch strings.TrimSpace(line) {
		case "Y", "y", "":
			break action

		case "Q", "q":
			return nil

		case "V", "v":
			viewDetails(scan, checkser.Invalid)
		case "A", "a":
			viewDetails(scan, checkser.Added)
		case "R", "r":
			viewDetails(scan, checkser.Removed)
		case "C", "c":
			viewDetails(scan, checkser.Changed)
		case "N", "n":
			viewDetails(scan, checkser.NoChange)
		case "F", "f":
			viewDetails(scan, checkser.Failed)
		}
	}

	return nil
}
