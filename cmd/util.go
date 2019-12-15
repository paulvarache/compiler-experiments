package cmd

import (
	"os"
	"path/filepath"
)

// ResolvePath returns the absolute path for a provided relative or avsolute path
// If relative will resolve from the current working directory
// All paths will be cleaned
func ResolvePath(relPath string) (string, error) {
	if filepath.IsAbs(relPath) {
		return filepath.Clean(relPath), nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", nil
	}
	abs, err := filepath.Abs(filepath.Join(cwd, relPath))
	if err != nil {
		return "", nil
	}
	return abs, nil
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
