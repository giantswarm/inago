// Package filesystemfake implements a fiile system that operates against in
// memory content.
package filesystemfake

import (
	"os"
	"path/filepath"

	"github.com/giantswarm/inago/file-system/spec"
)

// NewFileSystem creates a new fake filesystem. Operations are made against in
// memory content.
func NewFileSystem() filesystemspec.FileSystem {
	newFileSystem := &fake{
		Storages: map[string]map[string][]byte{},
	}

	return newFileSystem
}

type fake struct {
	Storages map[string]map[string][]byte
}

func (f *fake) ReadDir(dirname string) ([]os.FileInfo, error) {
	newFileInfos := []os.FileInfo{}

	if s, ok := f.Storages[dirname]; ok {
		for filename, content := range s {
			newFileInfos = append(newFileInfos, newFileInfo(filename, content))
		}
	}

	if len(newFileInfos) > 0 {
		return newFileInfos, nil
	}

	pathErr := &os.PathError{
		Op:   "open",
		Path: dirname,
		Err:  noSuchFileOrDirectoryError,
	}

	return nil, maskAny(pathErr)
}

func (f *fake) ReadFile(filename string) ([]byte, error) {
	dir := filepath.Dir(filename)
	base := filepath.Base(filename)

	if s, ok := f.Storages[dir]; ok {
		if bytes, ok := s[base]; ok {
			return bytes, nil
		}
	}

	pathErr := &os.PathError{
		Op:   "open",
		Path: filename,
		Err:  noSuchFileOrDirectoryError,
	}

	return nil, maskAny(pathErr)
}

func (f *fake) WriteFile(filename string, bytes []byte, perm os.FileMode) error {
	dir := filepath.Dir(filename)
	base := filepath.Base(filename)

	if _, ok := f.Storages[dir]; !ok {
		f.Storages[dir] = map[string][]byte{}
	}

	f.Storages[dir][base] = bytes
	return nil
}
