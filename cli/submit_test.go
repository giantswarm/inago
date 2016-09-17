package cli

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/juju/errgo"
)

func TestCreateSubmitRequest(t *testing.T) {
	tests := []struct {
		GroupDir      string
		Scale         int
		Files         []fileDesc
		Error         error
		DesiredSlices int
	}{
		{
			GroupDir: "g1",
			Scale:    1,
			Files: []fileDesc{
				{Path: "g1/g1-1.service", Content: "unit1"},
			},
			Error:         nil,
			DesiredSlices: 1,
		},
		{
			GroupDir: "g2",
			Scale:    3,
			Files: []fileDesc{
				{Path: "g2/g2-1.service", Content: "unit1"},
			},
			Error:         invalidScaleError,
			DesiredSlices: 0,
		},
		{
			GroupDir: "g3",
			Scale:    3,
			Files: []fileDesc{
				{Path: "g3/g3-1@.service", Content: "unit1"},
			},
			Error:         nil,
			DesiredSlices: 3,
		},
	}

	cleanDir := chdirTmp(t)
	defer cleanDir()

	for i, tt := range tests {
		prepareDir(t, i, filepath.Clean(tt.GroupDir), tt.Files)
		req, err := createSubmitRequest(tt.GroupDir, tt.Scale)

		if !reflect.DeepEqual(errgo.Cause(err), tt.Error) {
			t.Errorf("#%d: unexpected error = %v", i, err)
		}
		if tt.Error != nil {
			continue
		}
		if req.DesiredSlices != tt.DesiredSlices {
			t.Errorf("#%d: DesiredSlices = %d, want %d", i, req.DesiredSlices, tt.DesiredSlices)
		}
		if req.SliceIDs != nil {
			t.Errorf("#%d: SliceIDs = %v, want nil", i, req.SliceIDs)
		}
	}
}
