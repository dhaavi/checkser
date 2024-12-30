package checkser

import "fmt"

func (scan *Scan) DigestFiles() {
	scan.digest(scan.rootSum)
}

func (scan *Scan) digest(cs *Checksums) {
	stats := scan.Stats

files:
	for _, file := range cs.Files {
		// Check if a digest is needed.
		switch file.Change {
		case Removed, Failed:
			// Never digest.
		case Added, Changed, TimestampChanged:
			// Always digest.
		case NoChange:
			// Only digest if digest all is enabled.
			if !scan.cfg.DigestAll {
				stats.DigestSkipped.Add(1)
				stats.notify()
				continue files
			}
		}

		// Register file digest.
		stats.DigestFiles.Add(1)
		stats.notify()

		// Get existing hash or use default.
		h := Hash(file.Algorithm)
		if !h.IsValid() {
			h = scan.cfg.DefaultHash
		}

		// Check if the default hash is being forced.
		if scan.cfg.Rebuild {
			h = scan.cfg.DefaultHash
		}

		// Digest file.
		sum, err := h.DigestFile(file.Path)
		if err != nil {
			file.Change = Failed
			file.ErrMsgs = append(file.ErrMsgs, fmt.Sprintf("digest failed: %s", err))

			stats.DigestErrors.Add(1)
			stats.notify()

			continue files
		}

		// Write new hash sum to file.
		file.Changed.Algorithm = string(h)
		file.Changed.Digest = sum

		// Update change type.
		if file.Change != Added {
			switch {
			case file.Algorithm != file.Changed.Algorithm:
				file.Change = Changed
			case file.Digest != file.Changed.Digest:
				file.Change = Changed
			}
		}
	}

	for _, dir := range cs.Directories {
		// Check if a digest is needed.
		switch dir.Change {
		case Removed, Failed:
			// Never digest.
		default:
			// Digest everything else.
			scan.digest(dir.Checksums)
		}
	}
}
