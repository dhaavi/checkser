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
	cfg ScanConfig

	rootDir string
	rootSum *Checksums

	updatedAt time.Time
	updatedBy string

	Stats     *Stats
	writeErrs []string
}

type ScanConfig struct {
	// DefaultHash sets the default hash to use for new files.
	DefaultHash Hash

	// Rebuild completely rebuilds all checksums.
	// All files are digested.
	// All checksum files rewritten.
	Rebuild bool

	// DigestAll forces all files to be digested.
	// By default only files that have changed in size or modification time are digested.
	DigestAll bool

	// LiveUpdates enabled live update signalling using LiveUpdateSignal().
	// As stats are atomic there might inconsistencies during operation.
	LiveUpdates bool
}

func New(dir string, cfg ScanConfig) (*Scan, error) {
	// Get hostname for updated by.
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	// Check config.
	switch {
	case string(cfg.DefaultHash) == "":
		cfg.DefaultHash = DefaultHash
	case !cfg.DefaultHash.IsValid():
		return nil, ErrInvalidHashAlg
	}
	if cfg.Rebuild {
		cfg.DigestAll = true
	}

	// Create new scan.
	scan := &Scan{
		cfg:       cfg,
		rootDir:   dir,
		updatedAt: time.Now().Round(time.Second),
		updatedBy: hostname,
		Stats: &Stats{
			live: cfg.LiveUpdates,
		},
	}

	// Init live signal.
	if scan.Stats.live {
		scan.Stats.signal = make(chan struct{}, 1)
	}

	return scan, nil
}

func (scan *Scan) Scan() error {
	// Scan root dir.
	cs, err := scan.dir(scan.rootDir, nil)
	if err != nil {
		return err
	}
	scan.rootSum = cs

	// Scan iteratively from here.
	scan.dirs(cs)

	return nil
}

func (scan *Scan) dirs(cs *Checksums) {
	for _, dir := range cs.Directories {
		cs, err := scan.dir(dir.Path, dir)
		if err != nil {
			dir.Change = Failed
			dir.ErrMsgs = append(dir.ErrMsgs, fmt.Sprintf("failed to scan dir: %s", err))
			scan.Stats.FindingErrors.Add(1)
			scan.Stats.notify()
		} else {
			dir.Checksums = cs

			// Scan next level.
			scan.dirs(cs)
		}
	}
}

func (scan *Scan) dir(path string, pathDir *Directory) (*Checksums, error) {
	stats := scan.Stats

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
	if idx >= 0 {
		stats.FoundChecksums.Add(1)

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
				pathDir.writeChecksums = true // Force re-write.

				stats.FindingErrors.Add(1)
				stats.notify()
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
			stats.FoundDirs.Add(1)
			stats.notify()

			dir := cs.GetDir(entry.Name())
			if dir == nil {
				cs.AddDir(&Directory{
					Name:           entry.Name(),
					Path:           filepath.Join(path, entry.Name()),
					Change:         Added,
					writeChecksums: true, // Force write flag on new dirs.
				})
			} else {
				dir.Path = filepath.Join(path, entry.Name())
				dir.Change = NoChange
			}

		case entry.Type().IsRegular():
			stats.FoundFiles.Add(1)
			stats.notify()

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

					stats.FindingErrors.Add(1)
					stats.notify()
				} else {
					cs.AddFile(&File{
						Name:   entry.Name(),
						Path:   filepath.Join(path, entry.Name()),
						Change: Added,
						Changed: struct {
							Size      int64
							Modified  time.Time
							Algorithm string
							Digest    string
						}{
							Size:     info.Size(),
							Modified: info.ModTime(),
						},
					})
				}
			} else {
				// Gather Info
				info, err := entry.Info()
				if err != nil {
					file.Path = filepath.Join(path, entry.Name())
					file.Change = Failed
					file.ErrMsgs = []string{fmt.Sprintf("failed to get file info: %s", err)}

					stats.FindingErrors.Add(1)
					stats.notify()
				} else {
					file.Path = filepath.Join(path, entry.Name())
					file.AddChanges(info.Size(), info.ModTime())
				}
			}

		default:
			stats.FoundSpecial.Add(1)
			stats.notify()

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

					stats.FindingErrors.Add(1)
					stats.notify()
				} else {
					cs.AddSpecialFile(&Special{
						Name:   entry.Name(),
						Path:   filepath.Join(path, entry.Name()),
						Change: Added,
						Changed: struct {
							Type     string
							Modified time.Time
						}{
							Type:     specialType,
							Modified: info.ModTime(),
						},
					})
				}
			} else {
				// Gather Info
				info, err := entry.Info()
				if err != nil {
					specialFile.Path = filepath.Join(path, entry.Name())
					specialFile.Change = Failed
					specialFile.ErrMsgs = []string{fmt.Sprintf("failed to get file info: %s", err)}

					stats.FindingErrors.Add(1)
					stats.notify()
				} else {
					specialFile.Path = filepath.Join(path, entry.Name())
					specialFile.AddChanges(specialType, info.ModTime())
				}
			}
		}
	}

	cs.CheckMissing(path)
	return cs, nil
}

func (scan *Scan) WriteErrors() []string {
	return scan.writeErrs
}
