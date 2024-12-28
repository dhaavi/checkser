package checkser

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"time"
)

var ChecksumFilename = ".checkser.yml"

type Scan struct {
	rootDir string
	rootSum *Checksums

	updatedAt time.Time
	updatedBy string

	Stats struct {
		FoundDirs    uint64
		FoundFiles   uint64
		FoundSpecial uint64

		Errors uint64
	}
}

func ScanDir(dir string) (*Scan, error) {
	// Get hostname for updated by.
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	// Create new scan.
	scan := &Scan{
		rootDir:   dir,
		updatedAt: time.Now(),
		updatedBy: hostname,
	}

	// Scan root dir.
	cs, err := scan.dir(scan.rootDir, nil)
	if err != nil {
		return nil, err
	}
	scan.rootSum = cs

	// Scan iteratively from here.
	scan.dirs(cs)

	return scan, nil
}

func (scan *Scan) dirs(cs *Checksums) {
	for _, dir := range cs.Directories {
		cs, err := scan.dir(dir.Path, dir)
		if err != nil {
			dir.ErrMsgs = append(dir.ErrMsgs, fmt.Sprintf("failed to scan dir: %s", err))
			scan.Stats.Errors++
		} else {
			dir.Checksums = cs

			// Scan next level.
			scan.dirs(cs)
		}
	}
}

func (scan *Scan) dir(path string, pathDir *Directory) (*Checksums, error) {
	// Read dir.
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	// Load or create checksum file.
	var cs *Checksums
	idx := slices.IndexFunc(entries, func(entry os.DirEntry) bool {
		return entry.Name() == ChecksumFilename
	})
	if idx > 0 {
		// Load checksum file from dir.
		checksumData, err := os.ReadFile(filepath.Join(path, entries[idx].Name()))
		if err != nil {
			return nil, fmt.Errorf("failed to read checksum file %s: %w", entries[idx].Name(), err)
		}
		cs, err = LoadChecksums(checksumData)
		if err != nil {
			return nil, err
		}

		// If we have a path dir, check if the checksum matches.
		if pathDir != nil && pathDir.Algorithm != "" {
			dirChecksum, err := Hash(pathDir.Algorithm).Digest(checksumData)
			if err != nil {
				return nil, err
			}
			if dirChecksum == pathDir.Digest {
				pathDir.Verified = true
			} else {
				pathDir.ErrMsgs = append(pathDir.ErrMsgs, "dir integrity violated: checksum did not match, possibly checkser was used only on subset of data")
				scan.Stats.Errors++
			}
		}

	} else {
		// Create new checksum for this dir.
		cs = &Checksums{
			Version: 1,
		}
	}

	// Go through die entries and collect info.
	for _, entry := range entries {
		switch {
		case entry.Name() == ChecksumFilename:
			// Ignore checksum file itself.

		case entry.IsDir():
			scan.Stats.FoundDirs++

			dir := cs.GetDir(entry.Name())
			if dir == nil {
				cs.AddDir(&Directory{
					Name:   entry.Name(),
					Path:   filepath.Join(path, entry.Name()),
					Change: Added,
				})
			} else {
				dir.Change = NoChange
			}

		case entry.Type().IsRegular():
			scan.Stats.FoundFiles++

			file := cs.GetFile(entry.Name())
			if file == nil {
				// Gather Info
				info, err := entry.Info()
				if err != nil {
					cs.AddFile(&File{
						Name:    entry.Name(),
						Path:    filepath.Join(path, entry.Name()),
						Change:  Failed,
						ErrMsgs: []string{fmt.Sprintf("failed to get file info: %s", err)},
					})
					scan.Stats.Errors++
				} else {
					cs.AddFile(&File{
						Name:     entry.Name(),
						Path:     filepath.Join(path, entry.Name()),
						Size:     info.Size(),
						Modified: info.ModTime(),
						Change:   Added,
					})
				}
			} else {
				// Gather Info
				info, err := entry.Info()
				if err != nil {
					file.Change = Failed
					file.ErrMsgs = []string{fmt.Sprintf("failed to get file info: %s", err)}
					scan.Stats.Errors++
				} else {
					file.AddChanges(info.Size(), info.ModTime())
				}
			}

		default:
			scan.Stats.FoundSpecial++

			// Get special type.
			var specialType string
			switch {
			case entry.Type()&fs.ModeSymlink != 0:
				specialType = "symlink"
			case entry.Type()&fs.ModeNamedPipe != 0:
				specialType = "pipe"
			case entry.Type()&fs.ModeSocket != 0:
				specialType = "socket"
			case entry.Type()&fs.ModeDevice != 0:
				specialType = "device"
			case entry.Type()&fs.ModeCharDevice != 0:
				specialType = "chardevice"
			default:
				specialType = "other"
			}

			specialFile := cs.GetSpecialFile(entry.Name())
			if specialFile == nil {
				// Gather Info
				info, err := entry.Info()
				if err != nil {
					cs.AddSpecialFile(&Special{
						Name:    entry.Name(),
						Path:    filepath.Join(path, entry.Name()),
						Change:  Failed,
						ErrMsgs: []string{fmt.Sprintf("failed to get file info: %s", err)},
					})
					scan.Stats.Errors++
				} else {
					cs.AddSpecialFile(&Special{
						Name:     entry.Name(),
						Path:     filepath.Join(path, entry.Name()),
						Type:     specialType,
						Modified: info.ModTime(),
						Change:   Added,
					})
				}
			} else {
				// Gather Info
				info, err := entry.Info()
				if err != nil {
					specialFile.Change = Failed
					specialFile.ErrMsgs = []string{fmt.Sprintf("failed to get file info: %s", err)}
					scan.Stats.Errors++
				} else {
					specialFile.AddChanges(specialType, info.ModTime())
				}
			}
		}
	}

	return cs, nil
}
