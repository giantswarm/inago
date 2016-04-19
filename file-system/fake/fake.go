// Package filesystemfake implements a fiile system that operates against in
// memory content.
package filesystemfake

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/giantswarm/inago/file-system/spec"
)

// NewFileSystem creates a new fake filesystem. Operations are made against in
// memory content.
func NewFileSystem() filesystemspec.FileSystem {
	newFileSystem := &fake{
		Storage: map[string]os.FileInfo{},
	}

	return newFileSystem
}

type fake struct {
	Storage map[string]os.FileInfo
}

func (f *fake) ReadDir(dirname string) ([]os.FileInfo, error) {
	newFileInfos := []os.FileInfo{}

	for filename, fi := range f.Storage {
		if filename == dirname {
			if fi.IsDir() {
				continue
			} else {
				return nil, os.NewSyscallError("readdirent", notADirectoryError)
			}
		}
		if !strings.HasPrefix(filename, dirname) {
			continue
		}
		fn := strings.Replace(filename, dirname+"/", "", 1)
		if fn == "" {
			continue
		}
		splitted := strings.Split(fn, string(filepath.Separator))
		if len(splitted) != 1 {
			continue
		}
		if c, ok := fi.(fileInfo); ok {
			c.File.Name = splitted[0]
			newFileInfos = append(newFileInfos, c)
			continue
		}

		return nil, maskAny(invalidImplementationError)
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
	if fi, ok := f.Storage[filename]; ok {
		if c, ok := fi.(fileInfo); ok {
			b := c.File.Buffer.Bytes()
			return b, nil
		}

		return nil, maskAny(invalidImplementationError)
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

	if dir != "." {
		var ps []string
		for _, d := range strings.Split(filepath.FromSlash(dir), string(filepath.Separator)) {
			ps = append(ps, d)
			p := filepath.Join(ps...)
			f.Storage[p] = newDirFileInfo(p)
		}
	}

	f.Storage[filename] = newFileFileInfo(filename, bytes, perm)

	return nil
}
