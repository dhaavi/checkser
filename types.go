package checkser

import (
	"slices"
	"time"
)

// ForceHash forces the use of a specific hash.
var ForceHash string

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

func (file *File) GetChangedDigest(force bool) error {
	h := DefaultHash
	switch {
	case file.Change == Added:
		// New File!
	case file.Change == Changed:
		h = Hash(file.Algorithm)
		// File changed.
	case file.Change == TimestampChanged:
		h = Hash(file.Algorithm)
		// At least the timestamp changed, so we need to check.
	case force:
		h = Hash(file.Algorithm)
		// Force update!
	default:
		// Update not necessary, just copy the digest.
		file.Changed.Algorithm = file.Algorithm
		file.Changed.Digest = file.Digest
		return nil
	}

	// Check if a specific hash is being forced.
	// Eg. to switch all files to that hash.
	if ForceHash != "" {
		h = Hash(ForceHash)
	}

	// Get new digest.
	digest, err := h.DigestFile(file.Path)
	if err != nil {
		return err
	}
	file.Changed.Algorithm = string(h)
	file.Changed.Digest = digest

	return nil
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

	Checksums *Checksums `json:"-" yaml:"-"`
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

type Change uint8

const (
	Removed Change = iota
	Added
	Changed
	TimestampChanged
	NoChange
	Failed
)
