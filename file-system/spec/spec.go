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

	// WriteFile is the equivalent to ioutil.WriteFile.
	WriteFile(filename string, bytes []byte, perm os.FileMode) error
}
