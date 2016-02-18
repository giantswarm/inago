// Package filesystemreal implements a fiile system that operates against the
// underlying os.
package filesystemreal

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/giantswarm/formica/file-system/spec"
)

// NewFileSystem creates a new real filesystem. Operations are made against the
// underlying os.
func NewFileSystem() filesystemspec.FileSystem {
	newFileSystem := &real{}

	return newFileSystem
}

type real struct{}

func (r *real) ReadDir(dirname string) ([]os.FileInfo, error) {
	fileInfos, err := ioutil.ReadDir(dirname)
	if err != nil {
		return nil, maskAny(err)
	}

	return fileInfos, nil
}

func (r *real) ReadFile(filename string) ([]byte, error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, maskAny(err)
	}

	return bytes, nil
}

func (r *real) WriteFile(filename string, bytes []byte, perm os.FileMode) error {
	dir := filepath.Dir(filename)
	err := os.MkdirAll(dir, os.FileMode(0755))
	if err != nil {
		return maskAny(err)
	}

	err = ioutil.WriteFile(filename, bytes, perm)
	if err != nil {
		return maskAny(err)
	}

	return nil
}
