package checkser

import "sync/atomic"

type Stats struct {
	FoundDirs     atomic.Uint64
	FoundFiles    atomic.Uint64
	FoundSpecial  atomic.Uint64
	FindingErrors atomic.Uint64

	DigestFiles   atomic.Uint64
	DigestSkipped atomic.Uint64
	DigestErrors  atomic.Uint64

	// Changes
	Files   ChangeSet
	Dirs    ChangeSet
	Special ChangeSet
	Total   ChangeSet

	WriteChecksums atomic.Uint64
	WriteErrors    atomic.Uint64

	live   bool
	signal chan struct{}
}

type ChangeSet struct {
	Removed          atomic.Uint64
	Added            atomic.Uint64
	Changed          atomic.Uint64
	TimestampChanged atomic.Uint64
	NoChange         atomic.Uint64
	Failed           atomic.Uint64
}

func (scan *Scan) CalculateChangeStats() {
	scan.calcStats(scan.rootSum)
}

func (scan *Scan) calcStats(cs *Checksums) {
	stats := scan.Stats

	for _, file := range cs.Files {
		switch file.Change {
		case Removed:
			stats.Files.Removed.Add(1)
			stats.Total.Removed.Add(1)
		case Added:
			stats.Files.Added.Add(1)
			stats.Total.Added.Add(1)
		case Changed:
			stats.Files.Changed.Add(1)
			stats.Total.Changed.Add(1)
		case TimestampChanged:
			stats.Files.TimestampChanged.Add(1)
			stats.Total.TimestampChanged.Add(1)
		case NoChange:
			stats.Files.NoChange.Add(1)
			stats.Total.NoChange.Add(1)
		case Failed:
			stats.Files.Failed.Add(1)
			stats.Total.Failed.Add(1)
		}
	}

	for _, dir := range cs.Directories {
		switch dir.Change {
		case Removed:
			stats.Dirs.Removed.Add(1)
			stats.Total.Removed.Add(1)
		case Added:
			stats.Dirs.Added.Add(1)
			stats.Total.Added.Add(1)
		case Changed:
			stats.Dirs.Changed.Add(1)
			stats.Total.Changed.Add(1)
		case TimestampChanged:
			stats.Dirs.TimestampChanged.Add(1)
			stats.Total.TimestampChanged.Add(1)
		case NoChange:
			stats.Dirs.NoChange.Add(1)
			stats.Total.NoChange.Add(1)
		case Failed:
			stats.Dirs.Failed.Add(1)
			stats.Total.Failed.Add(1)
		}
	}

	for _, special := range cs.Specials {
		switch special.Change {
		case Removed:
			stats.Special.Removed.Add(1)
			stats.Total.Removed.Add(1)
		case Added:
			stats.Special.Added.Add(1)
			stats.Total.Added.Add(1)
		case Changed:
			stats.Special.Changed.Add(1)
			stats.Total.Changed.Add(1)
		case TimestampChanged:
			stats.Special.TimestampChanged.Add(1)
			stats.Total.TimestampChanged.Add(1)
		case NoChange:
			stats.Special.NoChange.Add(1)
			stats.Total.NoChange.Add(1)
		case Failed:
			stats.Special.Failed.Add(1)
			stats.Total.Failed.Add(1)
		}
	}

	for _, dir := range cs.Directories {
		if dir.Checksums != nil {
			scan.calcStats(dir.Checksums)
		}
	}
}
