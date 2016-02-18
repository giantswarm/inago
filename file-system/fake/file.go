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
	Mode    os.FileMode
	ModTime time.Time
	Buffer  *bytes.Reader
}

// newFile creates a new file instance based on a file name and it's contents
// from strings. The content of the new instance will be stored in an internal
// bytes.Reader instance. The default file mode will be 0777 and the last
// modification time the moment when the function is called.
func newFile(name string, content []byte) file {
	return file{
		Name:    name,
		Mode:    os.FileMode(0777),
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

// ReadAt wraps io.ReaderAt's functionality around the internalt bytes.Reader
// instance.
func (f file) ReadAt(p []byte, off int64) (n int, err error) {
	return f.Buffer.ReadAt(p, off)
}

// Seek wraps io.Seeker's functionality around the internalt bytes.Reader
// instance.
func (f file) Seek(offset int64, whence int) (int64, error) {
	return f.Buffer.Seek(offset, whence)
}

// Stat returns the fileInfo structure describing the file instance.
func (f file) Stat() (os.FileInfo, error) {
	return fileInfo{File: f}, nil
}

func newFileInfo(name string, content []byte) os.FileInfo {
	newFileInfo := fileInfo{
		File: newFile(name, content),
	}

	return newFileInfo
}

// fileInfo describes a wrapped file isntance and is returned by file.Stat
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

// IsDir always return false since it only uses file instances.
func (fi fileInfo) IsDir() bool {
	return false
}

// Sys always returns nil to stay conformant to the os.FileInfo interface.
func (fi fileInfo) Sys() interface{} {
	return nil
}
