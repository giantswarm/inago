package cli

import (
	"io/ioutil"
	"os"
	"testing"

	. "github.com/onsi/gomega"
)

func TestCreateSubmitRequest_NoSlices(t *testing.T) {
	RegisterTestingT(t)

	group := tempDir(t)
	defer os.RemoveAll(group)
	writeFile(t, group+"/"+group+"-1.service", "some content")
	scale := 1

	req, err := createSubmitRequest(group, scale)
	Expect(err).To(BeNil())
	Expect(req.Group).To(Equal(group))
	Expect(req.DesiredSlices).To(Equal(1))
	Expect(req.SliceIDs).To(BeNil())
}

func TestCreateSubmitRequest_NoSlices_InvalidScale(t *testing.T) {
	RegisterTestingT(t)

	group := tempDir(t)
	defer os.RemoveAll(group)
	writeFile(t, group+"/"+group+"-1.service", "some content")
	scale := 3

	_, err := createSubmitRequest(group, scale)
	Expect(err).To(Not(BeNil()))
}

func TestCreateSubmitRequest_WithSlices(t *testing.T) {
	RegisterTestingT(t)

	group := tempDir(t)
	defer os.RemoveAll(group)
	writeFile(t, group+"/"+group+"-1@.service", "some content")
	scale := 3

	req, err := createSubmitRequest(group, scale)
	Expect(err).To(BeNil())
	Expect(req.Group).To(Equal(group))
	Expect(req.DesiredSlices).To(Equal(3))
	Expect(len(req.SliceIDs)).To(Equal(0))
}

func tempDir(t *testing.T) string {
	name, err := ioutil.TempDir(".", "tmp-test-")
	if err != nil {
		t.Fatalf("unexpected TempDir error = %v", err)
	}
	return name
}

func writeFile(t *testing.T, path, content string) {
	if err := ioutil.WriteFile(path, []byte(content), os.FileMode(0664)); err != nil {
		t.Fatalf("unexpected WriteFile error = %v", err)
	}
}
