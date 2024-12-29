package checkser

import "fmt"

func (scan *Scan) DigestFiles(quick bool) {
	scan.digest(scan.rootSum, quick)
}

func (scan *Scan) digest(cs *Checksums, quick bool) {
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
			// Only digest if not in quick mode.
			if quick {
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
			h = DefaultHash
		}

		// Digest file.
		sum, err := h.DigestFile(file.Path)
		if err != nil {
			file.Change = Failed
			file.ErrMsgs = append(file.ErrMsgs, fmt.Sprintf("digest failed: %s", err))

			stats.DigestErrors.Add(1)
			stats.notify()
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
			scan.digest(dir.Checksums, quick)
		}
	}
}
