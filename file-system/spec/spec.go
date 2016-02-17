package filesystemspec

import (
	"os"
)

type FileSystem interface {
	ReadDir(dirname string) ([]os.FileInfo, error)
	ReadFile(filename string) ([]byte, error)
	WriteFile(filename string, bytes []byte, perm os.FileMode) error
}
