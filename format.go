package checkser

import "fmt"

func (scan *Scan) FmtFindStatus() []string {
	lines := make([]string, 5)
	lines[0] = fmt.Sprintf("Found Dirs: %d", scan.Stats.FoundDirs.Load())
	lines[1] = fmt.Sprintf("Found Files: %d", scan.Stats.FoundFiles.Load())
	lines[2] = fmt.Sprintf("Found Special: %d", scan.Stats.FoundSpecial.Load())
	lines[3] = fmt.Sprintf("Found Checksums: %d", scan.Stats.FoundChecksums.Load())
	lines[4] = fmt.Sprintf("Errors: %d", scan.Stats.FindingErrors.Load())
	return lines
}

func (scan *Scan) FmtFindStatusProgress() string {
	return fmt.Sprintf(
		"Scanning... (found %d dirs, %d files, %d other, %d checksums with %d errors)",
		scan.Stats.FoundDirs.Load(),
		scan.Stats.FoundFiles.Load(),
		scan.Stats.FoundSpecial.Load(),
		scan.Stats.FoundChecksums.Load(),
		scan.Stats.FindingErrors.Load(),
	)
}

func (scan *Scan) FmtDigestStatus() []string {
	lines := make([]string, 3)
	lines[0] = fmt.Sprintf("Digested Files: %d", scan.Stats.DigestFiles.Load())
	lines[1] = fmt.Sprintf("Skipped Files: %d", scan.Stats.DigestSkipped.Load())
	lines[2] = fmt.Sprintf("Errors: %d", scan.Stats.DigestErrors.Load())
	return lines
}

func (scan *Scan) FmtDigestStatusProgress() string {
	return fmt.Sprintf(
		"Digesting... (%.0f%%, %d/%d digested, %d skipped with %d errors)",
		float64(scan.Stats.DigestFiles.Load())*100/float64(scan.Stats.FoundFiles.Load()-scan.Stats.DigestSkipped.Load()),
		scan.Stats.DigestFiles.Load(),
		scan.Stats.FoundFiles.Load()-scan.Stats.DigestSkipped.Load(),
		scan.Stats.DigestSkipped.Load(),
		scan.Stats.DigestErrors.Load(),
	)
}

func (scan *Scan) FmtChangeStatus() []string {
	lines := make([]string, 6)
	lines[0] = fmt.Sprintf(
		"Removed: %d (%d files, %d dirs, %d other)",
		scan.Stats.Total.Removed.Load(),
		scan.Stats.Files.Removed.Load(),
		scan.Stats.Dirs.Removed.Load(),
		scan.Stats.Special.Removed.Load(),
	)
	lines[1] = fmt.Sprintf(
		"Added: %d (%d files, %d dirs, %d other)",
		scan.Stats.Total.Added.Load(),
		scan.Stats.Files.Added.Load(),
		scan.Stats.Dirs.Added.Load(),
		scan.Stats.Special.Added.Load(),
	)
	lines[2] = fmt.Sprintf(
		"Changed: %d (%d files, %d dirs, %d other)",
		scan.Stats.Total.Changed.Load(),
		scan.Stats.Files.Changed.Load(),
		scan.Stats.Dirs.Changed.Load(),
		scan.Stats.Special.Changed.Load(),
	)
	lines[3] = fmt.Sprintf(
		"TimestampChanged: %d (%d files, %d dirs, %d other)",
		scan.Stats.Total.TimestampChanged.Load(),
		scan.Stats.Files.TimestampChanged.Load(),
		scan.Stats.Dirs.TimestampChanged.Load(),
		scan.Stats.Special.TimestampChanged.Load(),
	)
	lines[4] = fmt.Sprintf(
		"NoChange: %d (%d files, %d dirs, %d other)",
		scan.Stats.Total.NoChange.Load(),
		scan.Stats.Files.NoChange.Load(),
		scan.Stats.Dirs.NoChange.Load(),
		scan.Stats.Special.NoChange.Load(),
	)
	lines[5] = fmt.Sprintf(
		"Failed: %d (%d files, %d dirs, %d other)",
		scan.Stats.Total.Failed.Load(),
		scan.Stats.Files.Failed.Load(),
		scan.Stats.Dirs.Failed.Load(),
		scan.Stats.Special.Failed.Load(),
	)
	return lines
}

func (scan *Scan) FmtWriteStatus() []string {
	lines := make([]string, 2)
	lines[0] = fmt.Sprintf("Written: %d", scan.Stats.WriteDone.Load())
	lines[1] = fmt.Sprintf("Errors: %d", scan.Stats.WriteErrors.Load())
	return lines
}

func (scan *Scan) FmtWriteStatusProgress() string {
	return fmt.Sprintf(
		"Writing Checksum Files... (%.0f%%, %d/%d written with %d errors)",
		float64(scan.Stats.WriteDone.Load())*100/float64(scan.Stats.WriteToDo.Load()),
		scan.Stats.WriteDone.Load(),
		scan.Stats.WriteToDo.Load(),
		scan.Stats.WriteErrors.Load(),
	)
}

// Iterate over all files, dirs and other files.
// Arguments must not be modified.
func (scan *Scan) Iterate(
	fileFunc func(*File),
	dirFunc func(*Directory),
	specialFunc func(*Special),
) {
	scan.iter(scan.rootSum, fileFunc, dirFunc, specialFunc)
}

func (scan *Scan) iter(
	cs *Checksums,
	fileFunc func(*File),
	dirFunc func(*Directory),
	specialFunc func(*Special),
) {
	for _, file := range cs.Files {
		fileFunc(file)
	}
	for _, special := range cs.Specials {
		specialFunc(special)
	}
	for _, dir := range cs.Directories {
		dirFunc(dir)
		if dir.Checksums != nil {
			scan.iter(dir.Checksums, fileFunc, dirFunc, specialFunc)
		}
	}
}
