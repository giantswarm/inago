package cli

import (
	"reflect"
	"testing"

	"github.com/giantswarm/inago/controller"
	"github.com/juju/errgo"
)

func TestCreateSubmitRequest(t *testing.T) {
	tests := []struct {
		Group         string
		Scale         int
		Files         []fileDesc
		DesiredSlices int
		Units         []controller.Unit
		Error         error
	}{
		{
			Group:         "g1",
			Scale:         1,
			DesiredSlices: 1,
			Files: []fileDesc{
				{Path: "g1/g1-1.service", Content: "unit1"},
			},
			Units: []controller.Unit{
				{Name: "g1-1.service", Content: "unit1"},
			},
			Error: nil,
		},
		{
			Group:         "g2",
			Scale:         3,
			DesiredSlices: 0,
			Files: []fileDesc{
				{Path: "g2/g2-1.service", Content: "unit1"},
			},
			Units: []controller.Unit{},
			Error: invalidScaleError,
		},
		{
			Group:         "g3",
			Scale:         3,
			DesiredSlices: 3,
			Files: []fileDesc{
				{Path: "g3/g3-1@.service", Content: "unit1"},
			},
			Units: []controller.Unit{
				{Name: "g3-1@.service", Content: "unit1"},
			},
			Error: nil,
		},
	}

	cleanDir := chdirTmp(t)
	defer cleanDir()

	for i, tt := range tests {
		prepareDir(t, i+1, tt.Group, tt.Files)
		req, err := createSubmitRequest(tt.Group, tt.Scale)

		if !reflect.DeepEqual(errgo.Cause(err), tt.Error) {
			t.Errorf("#%d: unexpected error = %v", i, err)
		}
		if tt.Error != nil {
			continue
		}
		if req.Group != tt.Group {
			t.Errorf("#%d: Group = %s, want %s", i, req.Group, tt.Group)
		}
		if req.DesiredSlices != tt.DesiredSlices {
			t.Errorf("#%d: DesiredSlices = %d, want %d", i, req.DesiredSlices, tt.DesiredSlices)
		}
		wsliceIDs := []string(nil)
		if !reflect.DeepEqual(req.SliceIDs, wsliceIDs) {
			t.Errorf("#%d: SliceIDs = %v, want %v", i, req.SliceIDs, wsliceIDs)
		}
	}
}
