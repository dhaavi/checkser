package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"

	"github.com/dhaavi/checkser"
)

func viewDetails(scan *checkser.Scan, filter checkser.Change) {
	v := &viewer{
		filter: filter,
	}

	// Write to tmp file for less, if available.
	if lessIsAvailable() {
		tmpFile, err := os.CreateTemp("", "checkser-change-view-")
		if err != nil {
			fmt.Printf("failed to create tmp file for viewing with less (err=%s), printing to stdout instead\n", err)
		} else {
			v.writer = tmpFile
			v.filepath = tmpFile.Name()
			defer tmpFile.Close()
		}
	}

	// Always fall back to stdout.
	if v.writer == nil {
		v.writer = os.Stdout
		fmt.Fprintln(v.writer, "Details:")
		defer fmt.Fprintln(v.writer, "")
	}

	// Write view.
	scan.Iterate(v.formatFile, v.formatDir, v.formatSpecial)

	// View tmp file with less, if used.
	if v.filepath != "" {
		cmd := exec.Command(lessBin, v.filepath)
		err := cmd.Run()
		if err != nil {
			fmt.Printf("less exited with error: %s\n", err)
		}
	}
}

type viewer struct {
	writer   io.Writer
	filepath string

	filter checkser.Change
}

func (v *viewer) view(change checkser.Change) bool {
	switch change {
	case checkser.Invalid:
		// Filter is disabled.
		return true

	case checkser.TimestampChanged:
		// TimestampChanged is always included in Changed.
		return v.filter == checkser.Changed

	default:
		// Check of filter matches.
		return change == v.filter
	}
}

func (v *viewer) formatFile(file *checkser.File) {
	if !v.view(file.Change) {
		return
	}

	switch file.Change {
	case checkser.Removed, checkser.NoChange, checkser.Failed:
		fmt.Fprintf(v.writer, "%s %s\n", file.Change, file.Path)
	case checkser.Added:
		fmt.Fprintf(v.writer, "%s %s (%dB %s %s)\n", file.Change, file.Path, file.Changed.Size, file.Changed.Algorithm, file.Changed.Digest)
	case checkser.Changed:
		fmt.Fprintf(v.writer, "%s %s (%dB %s %s => %dB %s %s)\n", file.Change, file.Path, file.Size, file.Algorithm, file.Digest, file.Changed.Size, file.Changed.Algorithm, file.Changed.Digest)
	case checkser.TimestampChanged:
		fmt.Fprintf(v.writer, "%s %s (%s => %s)\n", file.Change, file.Path, file.Modified, file.Changed.Modified)
	}

	// Print any error messages.
	for _, msg := range file.ErrMsgs {
		fmt.Fprintln(v.writer, "        Error: "+msg)
	}
}

func (v *viewer) formatDir(dir *checkser.Directory) {
	if !v.view(dir.Change) {
		return
	}

	fmt.Fprintf(v.writer, "%s %s/\n", dir.Change, dir.Path)

	// Print any error messages.
	for _, msg := range dir.ErrMsgs {
		fmt.Fprintln(v.writer, "        Error: "+msg)
	}
}

func (v *viewer) formatSpecial(special *checkser.Special) {
	if !v.view(special.Change) {
		return
	}

	switch special.Change {
	case checkser.Removed, checkser.NoChange, checkser.Failed:
		fmt.Fprintf(v.writer, "%s %s\n", special.Change, special.Path)
	case checkser.Added:
		fmt.Fprintf(v.writer, "%s %s (%s)\n", special.Change, special.Path, special.Changed.Type)
	case checkser.Changed:
		fmt.Fprintf(v.writer, "%s %s (%s => %s)\n", special.Change, special.Path, special.Type, special.Changed.Type)
	case checkser.TimestampChanged:
		fmt.Fprintf(v.writer, "%s %s (%s => %s)\n", special.Change, special.Path, special.Modified, special.Changed.Modified)
	}

	// Print any error messages.
	for _, msg := range special.ErrMsgs {
		fmt.Fprintln(v.writer, "        Error: "+msg)
	}
}

var (
	lessBin           string
	lessBinSearchOnce sync.Once
)

func lessIsAvailable() bool {
	lessBinSearchOnce.Do(func() {
		path, err := exec.LookPath("less")
		if err == nil {
			lessBin = path
		}
	})

	return lessBin != ""
}
