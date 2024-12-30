package checkser

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
)

func (scan *Scan) WriteChecksumFiles() {
	scan.Stats.WriteToDo.Store(1) // Root Dir.
	scan.prepareForWriting(scan.rootSum)

	scan.writeChecksums(scan.rootDir, scan.rootSum)
}

func (scan *Scan) prepareForWriting(cs *Checksums) (writeChecksums bool) {

	// Prepare sub dirs.
	for _, dir := range cs.Directories {
		if dir.Checksums != nil {
			if scan.prepareForWriting(dir.Checksums) || scan.cfg.Rebuild {
				scan.Stats.WriteToDo.Add(1)
				dir.writeChecksums = true
				writeChecksums = true
			}
		}
	}

	// Update metadata.
	cs.UpdatedAt = scan.updatedAt
	cs.UpdatedBy = scan.updatedBy

	// Check if checksums need to be (re)written.
	for _, file := range cs.Files {
		switch file.Change {
		case NoChange, Failed:
			// Checksums update not necessary.
		default:
			writeChecksums = true
		}
	}
	for _, dir := range cs.Directories {
		switch dir.Change {
		case NoChange, Failed:
			// Checksums update not necessary.
		default:
			writeChecksums = true
		}
	}
	for _, special := range cs.Specials {
		switch special.Change {
		case NoChange, Failed:
			// Checksums update not necessary.
		default:
			writeChecksums = true
		}
	}

	// Purge unneeded entries.
	cs.Files = slices.DeleteFunc(cs.Files, func(file *File) bool {
		switch file.Change {
		case Added, Changed, TimestampChanged, NoChange:
			// Keep these entries.
			return false
		default:
			// Remove other entries.
			return true
		}
	})
	cs.Directories = slices.DeleteFunc(cs.Directories, func(dir *Directory) bool {
		switch dir.Change {
		case Added, Changed, TimestampChanged, NoChange:
			// Keep these entries.
			return false
		default:
			// Remove other entries.
			return true
		}
	})
	cs.Specials = slices.DeleteFunc(cs.Specials, func(special *Special) bool {
		switch special.Change {
		case Added, Changed, TimestampChanged, NoChange:
			// Keep these entries.
			return false
		default:
			// Remove other entries.
			return true
		}
	})

	// Apply changed data.
	for _, file := range cs.Files {
		switch file.Change {
		case Added, Changed, TimestampChanged:
			file.Size = file.Changed.Size
			file.Modified = file.Changed.Modified
			file.Algorithm = file.Changed.Algorithm
			file.Digest = file.Changed.Digest
		}
	}
	for _, special := range cs.Specials {
		switch special.Change {
		case Added, Changed, TimestampChanged:
			special.Type = special.Changed.Type
			special.Modified = special.Changed.Modified
		}
	}

	return writeChecksums
}

func (scan *Scan) writeChecksums(path string, cs *Checksums) (alg, sum string) {
	stats := scan.Stats
	defer stats.notify()

	// First write all sub dirs.
	for _, dir := range cs.Directories {
		if dir.writeChecksums {
			dir.Algorithm, dir.Digest = scan.writeChecksums(dir.Path, dir.Checksums)
		}
	}

	// Serialize checksums.
	packed, err := PackChecksums(cs)
	if err != nil {
		scan.writeErrs = append(scan.writeErrs, fmt.Sprintf("%s: serialization failed: %s", path, err))
		scan.Stats.WriteErrors.Add(1)
		return
	}

	// Write checksums file.
	err = os.WriteFile(filepath.Join(path, ChecksumFilename), packed, 0o0755)
	if err != nil {
		scan.writeErrs = append(scan.writeErrs, fmt.Sprintf("%s: write failed: %s", path, err))
		scan.Stats.WriteErrors.Add(1)
	} else {
		scan.Stats.WriteDone.Add(1)
	}

	// Digest for parent checksums.
	sum, err = scan.cfg.DefaultHash.Digest(packed)
	if err != nil {
		scan.writeErrs = append(scan.writeErrs, fmt.Sprintf("%s: hashing failed (non-critical): %s", path, err))
		scan.Stats.WriteErrors.Add(1)
		return "", ""
	}
	return string(scan.cfg.DefaultHash), sum
}
