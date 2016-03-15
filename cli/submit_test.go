package cli

import (
	"os"
	"testing"

	"github.com/giantswarm/inago/file-system/fake"
	"github.com/giantswarm/inago/file-system/spec"

	. "github.com/onsi/gomega"
)

func givenFileSystemWithSingleUnitGroup(name string) filesystemspec.FileSystem {
	fs := filesystemfake.NewFileSystem()
	fs.WriteFile(name+"/"+name+"-1.service", []byte(`some content`), os.FileMode(0644))
	return fs
}
func givenFileSystemWithSliceableUnitGroup(name string) filesystemspec.FileSystem {
	fs := filesystemfake.NewFileSystem()
	fs.WriteFile(name+"/"+name+"-1@.service", []byte(`some content`), os.FileMode(0644))
	return fs
}

func TestCreateSubmitRequest_NoSlices(t *testing.T) {
	RegisterTestingT(t)

	groupname := "foo"
	fs := givenFileSystemWithSingleUnitGroup(groupname)
	scale := 1

	req, err := createSubmitRequest(fs, groupname, scale)
	Expect(err).To(BeNil())
	Expect(req.Group).To(Equal(groupname))
	Expect(req.DesiredSlices).To(Equal(1))
	Expect(req.SliceIDs).To(BeNil())
}

func TestCreateSubmitRequest_NoSlices_InvalidScale(t *testing.T) {
	RegisterTestingT(t)

	groupname := "foo"
	fs := givenFileSystemWithSingleUnitGroup(groupname)
	scale := 3

	_, err := createSubmitRequest(fs, groupname, scale)
	Expect(err).To(Not(BeNil()))

}

func TestCreateSubmitRequest_WithSlices(t *testing.T) {
	RegisterTestingT(t)

	groupname := "foo"
	fs := givenFileSystemWithSliceableUnitGroup(groupname)

	scale := 3

	req, err := createSubmitRequest(fs, groupname, scale)
	Expect(err).To(BeNil())
	Expect(req.Group).To(Equal(groupname))
	Expect(req.DesiredSlices).To(Equal(3))
	Expect(len(req.SliceIDs)).To(Equal(0))
}
