package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dhaavi/checkser"
	"github.com/spf13/cobra"
)

var (
	runInteractive bool
	runUpdate      bool
	runVerify      bool
)

func run(cmd *cobra.Command, args []string) error {
	dir, err := filepath.Abs(args[0])
	if err != nil {
		return fmt.Errorf("invalid directory: %w", err)
	}

	// Scan the directory for checksums and files.
	fmt.Println("Finding files and directories...")
	scan, err := checkser.ScanDir(dir, checkser.ScanConfig{
		DefaultHash: checkser.Hash(flagDefaultHash),
		Rebuild:     flagRebuild,
		DigestAll:   flagDigestAll || runVerify,
	})
	if err != nil {
		return fmt.Errorf("invalid directory: %w", err)
	}
	for _, line := range scan.FmtFindStatus() {
		fmt.Println(line)
	}
	fmt.Println("")

	// Prompt before continuing when there are errors.
	cliReader := bufio.NewReader(os.Stdin)
	if !runInteractive && scan.Stats.FindingErrors.Load() > 0 {
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
				viewDetails(scan, checkser.ErrMsgs)
			}
		}
		fmt.Println("")
	}

	// Digest any needed files.
	fmt.Println("Digesting files...")
	scan.DigestFiles()
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

	// Check if there are any changes.
	switch {
	case scan.Stats.FindingErrors.Load() > 0:
	case scan.Stats.DigestErrors.Load() > 0:
	case scan.Stats.Total.Removed.Load() > 0:
	case scan.Stats.Total.Added.Load() > 0:
	case scan.Stats.Total.Changed.Load() > 0:
	case scan.Stats.Total.TimestampChanged.Load() > 0:
	case scan.Stats.Total.Failed.Load() > 0:
	default:
		fmt.Printf(
			"Checked all %d files, %d dirs and %d other. No changes found.\n",
			scan.Stats.Files.NoChange.Load(),
			scan.Stats.Dirs.NoChange.Load(),
			scan.Stats.Special.NoChange.Load(),
		)
		return nil
	}

	// Return an error if we are just verifying.
	if runVerify {
		return errors.New("changes or errors detected")
	}

	if runInteractive {
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
		fmt.Println("")
	}

	// Write checksum files.
	scan.WriteChecksumFiles()
	fmt.Printf("Successfully written %d checksum files.\n", scan.Stats.WriteDone.Load())
	if scan.Stats.WriteErrors.Load() > 0 {
		fmt.Printf("Encountered %d errors during writing checksum files:\n", scan.Stats.WriteErrors.Load())
		for _, line := range scan.WriteErrors() {
			fmt.Println(line)
		}
	}

	// Return an error if running update.
	if runUpdate {
		if scan.Stats.FindingErrors.Load() > 0 ||
			scan.Stats.DigestErrors.Load() > 0 ||
			scan.Stats.WriteErrors.Load() > 0 {
			fmt.Println("")
			return fmt.Errorf(
				"update complete, encountered %d scan errors, %d digest errors and %d write errors",
				scan.Stats.FindingErrors.Load(),
				scan.Stats.DigestErrors.Load(),
				scan.Stats.WriteErrors.Load(),
			)
		}
	}

	return nil
}
