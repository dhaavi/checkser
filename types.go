package checkser

import (
	"slices"
	"time"
)

type Checksums struct {
	Version int `json:"checkser,omitempty" yaml:"checkser,omitempty"`

	UpdatedAt time.Time `json:"updated_at,omitempty" yaml:"updated_at,omitempty"`
	UpdatedBy string    `json:"updated_by,omitempty" yaml:"updated_by,omitempty"`

	Files       []*File      `json:"files,omitempty" yaml:"files,omitempty"`
	Directories []*Directory `json:"dirs,omitempty" yaml:"dirs,omitempty"`
	Specials    []*Special   `json:"other,omitempty" yaml:"other,omitempty"`
}

type File struct {
	Name      string    `json:"name,omitempty" yaml:"name,omitempty"`
	Path      string    `json:"-" yaml:"-"`
	Size      int64     `json:"size,omitempty" yaml:"size,omitempty"`
	Modified  time.Time `json:"mod,omitempty" yaml:"mod,omitempty"`
	Algorithm string    `json:"alg,omitempty" yaml:"alg,omitempty"`
	Digest    string    `json:"sum,omitempty" yaml:"sum,omitempty"`

	Change  Change   `json:"-" yaml:"-"`
	ErrMsgs []string `json:"-" yaml:"-"`
	Changed struct {
		Size      int64
		Modified  time.Time
		Algorithm string
		Digest    string
	} `json:"-" yaml:"-"`
}

func (cs *Checksums) GetFile(name string) *File {
	idx := slices.IndexFunc(cs.Files, func(f *File) bool {
		return f.Name == name
	})
	if idx >= 0 {
		return cs.Files[idx]
	}
	return nil
}

func (cs *Checksums) AddFile(newFile *File) {
	cs.Files = append(cs.Files, newFile)
}

func (file *File) AddChanges(size int64, modified time.Time) {
	// Apply.
	file.Changed.Size = size
	file.Changed.Modified = modified

	// Check what kind of change it is when it was already seen.
	switch {
	case file.Size != size:
		file.Change = Changed
	case !file.Modified.Equal(modified):
		file.Change = TimestampChanged
	default:
		file.Change = NoChange
	}
}

type Directory struct {
	Name      string `json:"name,omitempty" yaml:"name,omitempty"`
	Path      string `json:"-" yaml:"-"`
	Algorithm string `json:"alg,omitempty" yaml:"alg,omitempty"`
	Digest    string `json:"sum,omitempty" yaml:"sum,omitempty"`

	Verified bool `json:"-" yaml:"-"`

	Change  Change   `json:"-" yaml:"-"`
	ErrMsgs []string `json:"-" yaml:"-"`
	Changed struct {
		ChangedAlgorithm string
		ChangedDigest    string
	} `json:"-" yaml:"-"`

	Checksums      *Checksums `json:"-" yaml:"-"`
	writeChecksums bool
}

func (cs *Checksums) GetDir(name string) *Directory {
	idx := slices.IndexFunc(cs.Directories, func(d *Directory) bool {
		return d.Name == name
	})
	if idx >= 0 {
		return cs.Directories[idx]
	}
	return nil
}

func (cs *Checksums) AddDir(newDir *Directory) {
	cs.Directories = append(cs.Directories, newDir)
}

type Special struct {
	Name     string    `json:"name,omitempty" yaml:"name,omitempty"`
	Path     string    `json:"-" yaml:"-"`
	Type     string    `json:"type,omitempty" yaml:"type,omitempty"`
	Modified time.Time `json:"mod,omitempty" yaml:"mod,omitempty"`

	Change  Change   `json:"-" yaml:"-"`
	ErrMsgs []string `json:"-" yaml:"-"`
	Changed struct {
		Type     string
		Modified time.Time
	} `json:"-" yaml:"-"`
}

func (cs *Checksums) GetSpecialFile(name string) *Special {
	idx := slices.IndexFunc(cs.Specials, func(s *Special) bool {
		return s.Name == name
	})
	if idx >= 0 {
		return cs.Specials[idx]
	}
	return nil
}

func (cs *Checksums) AddSpecialFile(newSpecialFile *Special) {
	cs.Specials = append(cs.Specials, newSpecialFile)
}

func (file *Special) AddChanges(specialType string, modified time.Time) {
	// Apply.
	file.Changed.Type = specialType
	file.Changed.Modified = modified

	// Check what kind of change it is when it was already seen.
	switch {
	case file.Type != specialType:
		file.Change = Changed
	case !file.Modified.Equal(modified):
		file.Change = TimestampChanged
	default:
		file.Change = NoChange
	}
}

type Change int8

const (
	ErrMsgs Change = -2
	Invalid Change = -1

	Removed Change = iota
	Added
	Changed
	TimestampChanged
	NoChange
	Failed
)

func (c Change) String() string {
	switch c {
	case Removed:
		return "removed"
	case Added:
		return "added"
	case Changed:
		return "changed"
	case TimestampChanged:
		return "timestamp changed"
	case NoChange:
		return "no change"
	case Failed:
		return "failed"
	default:
		return "unknown"
	}
}
