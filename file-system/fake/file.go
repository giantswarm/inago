package filesystemfake

import (
	"bytes"
	"os"
	"time"
)

// Fake represents are readable in memory version of os.File which can be used
// for testing.
type file struct {
	Name    string
	Dir     bool
	Mode    os.FileMode
	ModTime time.Time
	Buffer  *bytes.Reader
}

// newDir creates a new file instance based on a dir name. The default file
// mode will be 0644 and the last modification time the moment when the
// function is called.
func newDir(name string) file {
	return file{
		Name:    name,
		Dir:     true,
		Mode:    os.FileMode(0644),
		ModTime: time.Now(),
		Buffer:  nil,
	}
}

// newFile creates a new file instance based on a file name and it's contents
// from strings. The content of the new instance will be stored in an internal
// bytes.Reader instance. The last modification time the moment when the
// function is called.
func newFile(name string, content []byte, perm os.FileMode) file {
	return file{
		Name:    name,
		Dir:     false,
		Mode:    perm,
		ModTime: time.Now(),
		Buffer:  bytes.NewReader(content),
	}
}

// Close wraps io.Closer's functionality and does no operation.
func (f file) Close() error {
	return nil
}

// Read wraps io.Reader's functionality around the internal bytes.Reader
// instance.
func (f file) Read(p []byte) (n int, err error) {
	return f.Buffer.Read(p)
}

// ReadAt wraps io.ReaderAt's functionality around the internal bytes.Reader
// instance.
func (f file) ReadAt(p []byte, off int64) (n int, err error) {
	return f.Buffer.ReadAt(p, off)
}

// Seek wraps io.Seeker's functionality around the internal bytes.Reader
// instance.
func (f file) Seek(offset int64, whence int) (int64, error) {
	return f.Buffer.Seek(offset, whence)
}

// Stat returns the fileInfo structure describing the file instance.
func (f file) Stat() (os.FileInfo, error) {
	return fileInfo{File: f}, nil
}

func newFileFileInfo(name string, content []byte, perm os.FileMode) os.FileInfo {
	newFileInfo := fileInfo{
		File: newFile(name, content, perm),
	}

	return newFileInfo
}

func newDirFileInfo(name string) os.FileInfo {
	newFileInfo := fileInfo{
		File: newDir(name),
	}

	return newFileInfo
}

// fileInfo describes a wrapped file instance and is returned by file.Stat
type fileInfo struct {
	File file
}

// Name returns the base name of the file instance.
func (fi fileInfo) Name() string {
	return fi.File.Name
}

// Size returns the length in bytes of the file's internal bytes.Reader
// instance.
func (fi fileInfo) Size() int64 {
	return int64(fi.File.Buffer.Len())
}

// Mode returns file mode bits of the file instance.
func (fi fileInfo) Mode() os.FileMode {
	return fi.File.Mode
}

// ModTime returns the modification time of the file instance.
func (fi fileInfo) ModTime() time.Time {
	return fi.File.ModTime
}

// IsDir determines whether the current file represents a directory or a file.
func (fi fileInfo) IsDir() bool {
	return fi.File.Dir
}

// Sys always returns nil to stay conformant to the os.FileInfo interface.
func (fi fileInfo) Sys() interface{} {
	return nil
}
