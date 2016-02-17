// Package filesystemspec describes the interface of file system
// implementations.
package filesystemspec

import (
	"os"
)

// FileSystem implements file system operations.
type FileSystem interface {
	// ReadDir implements os.ReadDir.
	ReadDir(dirname string) ([]os.FileInfo, error)

	// ReadFile implements ioutil.ReadFile.
	ReadFile(filename string) ([]byte, error)

	// WriteFile implements ioutil.WriteFile.
	WriteFile(filename string, bytes []byte, perm os.FileMode) error
}
