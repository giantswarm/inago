// Package filesystemspec describes the interface of file system
// implementations.
package filesystemspec

import (
	"os"
)

// FileSystem implements file system operations.
type FileSystem interface {
	// ReadDir is the equivalent to os.ReadDir.
	ReadDir(dirname string) ([]os.FileInfo, error)

	// ReadFile is the equivalent to ioutil.ReadFile.
	ReadFile(filename string) ([]byte, error)

	// WriteFile is the equivalent to ioutil.WriteFile with one additional
	// behavior. It will automatically create a directory structure using
	// os.MkdirAll for the real implementation in case the file path provides a
	// nested file within a directory structure.
	WriteFile(filename string, bytes []byte, perm os.FileMode) error
}
