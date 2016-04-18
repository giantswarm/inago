package filesystemfake

import (
	"os"
	"testing"
)

type writeFiles struct {
	FileName string
	Bytes    []byte
	Perm     os.FileMode
}

func Test_FileSystem_ReadDir(t *testing.T) {
	testCases := []struct {
		WriteFiles   []writeFiles
		DirName      string
		Expected     string // expected dir name
		ErrorMatcher func(err error) bool
	}{
		{
			WriteFiles:   nil,
			DirName:      "",
			Expected:     "",
			ErrorMatcher: nil,
		},
		{
			WriteFiles:   nil,
			DirName:      "mydir",
			Expected:     "",
			ErrorMatcher: IsNoSuchFileOrDirectory,
		},
		{
			WriteFiles: []writeFiles{
				{
					FileName: "foo",
					Bytes:    []byte("bar"),
					Perm:     os.ModePerm,
				},
			},
			DirName:      "mydir",
			Expected:     "",
			ErrorMatcher: IsNoSuchFileOrDirectory,
		},
		{
			WriteFiles: []writeFiles{
				{
					FileName: "foo/bar",
					Bytes:    []byte("baz"),
					Perm:     os.ModePerm,
				},
			},
			DirName:      "mydir",
			Expected:     "",
			ErrorMatcher: IsNoSuchFileOrDirectory,
		},
		{
			WriteFiles: []writeFiles{
				{
					FileName: "mydir/foo/bar",
					Bytes:    []byte("baz"),
					Perm:     os.ModePerm,
				},
			},
			DirName:      "mydir",
			Expected:     "mydir",
			ErrorMatcher: nil,
		},
		{
			WriteFiles: []writeFiles{
				{
					FileName: "mydir/foo/bar.txt",
					Bytes:    []byte("baz"),
					Perm:     os.ModePerm,
				},
			},
			DirName:      "mydir",
			Expected:     "mydir",
			ErrorMatcher: nil,
		},
		{
			WriteFiles: []writeFiles{
				{
					FileName: "mydir/foo/bar.ext",
					Bytes:    []byte("baz"),
					Perm:     os.ModePerm,
				},
			},
			DirName:      "mydir",
			Expected:     "mydir",
			ErrorMatcher: nil,
		},
	}

	for i, testCase := range testCases {
		fs := NewFileSystem()

		for j, writeFile := range testCase.WriteFiles {
			err := fs.WriteFile(writeFile.FileName, writeFile.Bytes, writeFile.Perm)
			if err != nil {
				t.Fatal("case", j+1, "of", i+1, "expected", nil, "got", err)
			}
		}

		fileInfos, err := fs.ReadDir(testCase.DirName)
		if testCase.ErrorMatcher != nil && !testCase.ErrorMatcher(err) {
			t.Fatal("case", i+1, "expected", true, "got", false)
		}
		if testCase.ErrorMatcher == nil {
			for j, fileInfo := range fileInfos {
				if fileInfo.Name() != testCase.Expected {
					t.Fatal("case", j+1, "of", i+1, "expected", testCase.Expected, "got", fileInfo.Name())
				}
			}
		}
	}
}

func Test_FileSystem_ReadFile(t *testing.T) {
	testCases := []struct {
		WriteFiles   []writeFiles
		FileName     string
		Expected     string // expected file content
		ErrorMatcher func(err error) bool
	}{
		{
			WriteFiles:   nil,
			FileName:     "",
			Expected:     "",
			ErrorMatcher: nil,
		},
		{
			WriteFiles:   nil,
			FileName:     "myfile",
			Expected:     "",
			ErrorMatcher: IsNoSuchFileOrDirectory,
		},
		{
			WriteFiles: []writeFiles{
				{
					FileName: "foo",
					Bytes:    []byte("bar"),
					Perm:     os.ModePerm,
				},
			},
			FileName:     "myfile",
			Expected:     "",
			ErrorMatcher: IsNoSuchFileOrDirectory,
		},
		{
			WriteFiles: []writeFiles{
				{
					FileName: "foo/bar",
					Bytes:    []byte("baz"),
					Perm:     os.ModePerm,
				},
			},
			FileName:     "myfile",
			Expected:     "",
			ErrorMatcher: IsNoSuchFileOrDirectory,
		},
		{
			WriteFiles: []writeFiles{
				{
					FileName: "bar",
					Bytes:    []byte("123"),
					Perm:     os.ModePerm,
				},
			},
			FileName:     "bar",
			Expected:     "123",
			ErrorMatcher: nil,
		},
		{
			WriteFiles: []writeFiles{
				{
					FileName: "foo/bar.jpg",
					Bytes:    []byte("content"),
					Perm:     os.ModePerm,
				},
			},
			FileName:     "foo/bar.jpg",
			Expected:     "content",
			ErrorMatcher: nil,
		},
	}

	for i, testCase := range testCases {
		fs := NewFileSystem()

		for j, writeFile := range testCase.WriteFiles {
			err := fs.WriteFile(writeFile.FileName, writeFile.Bytes, writeFile.Perm)
			if err != nil {
				t.Fatal("case", j+1, "of", i+1, "expected", nil, "got", err)
			}
		}

		b, err := fs.ReadFile(testCase.FileName)
		if testCase.ErrorMatcher != nil && !testCase.ErrorMatcher(err) {
			t.Fatal("case", i+1, "expected", true, "got", false)
		}
		if testCase.ErrorMatcher == nil {
			if string(b) != testCase.Expected {
				t.Fatal("case", i+1, "expected", testCase.Expected, "got", string(b))
			}
		}
	}
}
