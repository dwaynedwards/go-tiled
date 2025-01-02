package tiled

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

var ResourcePath = ""

// New returns a Map from the given path
func New(path string) (*Map, error) {
	if path == "" {
		return nil, errors.New("file path is empty")
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open map file: %w", err)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			fmt.Printf("error closing map file handler %s", errors.Unwrap(err))
		}
	}(f)

	buf, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read map file: %w", err)
	}

	ResourcePath = filepath.Dir(path)
	var m Map
	err = xml.Unmarshal(buf, &m)
	if err != nil {
		return nil, fmt.Errorf("failed to parse map file: %w", err)
	}
	return &m, nil
}
