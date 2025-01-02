package checkser

import (
	"errors"

	"golang.org/x/text/unicode/norm"
	"gopkg.in/yaml.v3"
)

// Errors.
var (
	ErrInvalidChecksumFile = errors.New("invalid checksum file")
	ErrUnsupportedVersion  = errors.New("unsupported version")
)

func LoadChecksums(data []byte) (*Checksums, error) {
	// Load data into struct.
	cs := &Checksums{}
	err := yaml.Unmarshal(data, cs)
	if err != nil {
		return nil, err
	}

	// Check if file is correct, as far as we can tell.
	switch {
	case cs.Version > 1:
		return nil, ErrUnsupportedVersion
	case cs.Version <= 0:
		return nil, ErrInvalidChecksumFile
	}

	// Normalize names.
	for _, file := range cs.Files {
		file.Name = norm.NFC.String(file.Name)
	}
	for _, dir := range cs.Directories {
		dir.Name = norm.NFC.String(dir.Name)
	}
	for _, special := range cs.Specials {
		special.Name = norm.NFC.String(special.Name)
	}

	return cs, nil
}

func PackChecksums(cs *Checksums) ([]byte, error) {
	return yaml.Marshal(cs)
}
